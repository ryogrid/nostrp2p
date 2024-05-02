package core

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"os"
	"sync"
)

type LogFile struct {
	file        *os.File
	fileMtx     *sync.Mutex
	curSize     int64
	lastReadPos int64
}

type EventDataLogger struct {
	eventLogFile             *LogFile
	reSendNeededEvtLogFile   *LogFile
	reSendFinishedEvtLogFile *LogFile
}

func NewEventDataLogger(logFnameBase string) *EventDataLogger {
	return &EventDataLogger{
		eventLogFile:             newLogFile(logFnameBase + ".evtlog"),
		reSendNeededEvtLogFile:   newLogFile(logFnameBase + ".rsevtlog"),
		reSendFinishedEvtLogFile: newLogFile(logFnameBase + ".rsfevtlog"),
	}
}

func newLogFile(logFnameBase string) *LogFile {
	file, err := os.OpenFile(logFnameBase, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalln("can't open log file: " + logFnameBase)
		return nil
	}
	file.Seek(io.SeekStart, 0)
	return &LogFile{
		file:        file,
		fileMtx:     &sync.Mutex{},
		curSize:     -1,
		lastReadPos: 0,
	}
}

func (l *EventDataLogger) GetLogfileSize(lfile *LogFile) int64 {
	fi, err := lfile.file.Stat()
	if err != nil {
		panic(err)
	}
	return fi.Size()
}

func (l *EventDataLogger) WriteLog(lfile *LogFile, data []byte) error {
	lfile.fileMtx.Lock()
	defer lfile.fileMtx.Unlock()
	if lfile.curSize < 0 {
		fileInfo, err := lfile.file.Stat()
		if err != nil {
			return err
		}
		lfile.curSize = fileInfo.Size()
		lfile.file.Seek(lfile.curSize, 0)
	}
	sizeBuf := make([]byte, 0)
	sizeBuf = binary.BigEndian.AppendUint32(sizeBuf, uint32(len(data)))
	// each log entry is prefixed with a 4-byte size
	n, err := lfile.file.Write(sizeBuf)
	if err != nil || n != 4 {
		panic(err)
	}
	n, err = lfile.file.Write(data)
	if err != nil || n != len(data) {
		panic(err)
	}
	err = lfile.file.Sync()
	if err != nil {
		panic(err)
	}
	lfile.curSize += int64(n + 4)

	return err
}

// read a log entry
func (l *EventDataLogger) ReadLog(lfile *LogFile) (int, []byte, error) {
	lfile.fileMtx.Lock()
	defer lfile.fileMtx.Unlock()
	sizeBuf := make([]byte, 4)
	n, err := lfile.file.Read(sizeBuf)
	if err != nil || n != 4 {
		if errors.Is(err, io.EOF) {
			// we have reached the end of the file
			return -1, nil, err
		} else {
			// file broken (I/O error while writing log at last launch)
			// set lastReadPos to the next log writing point
			lfile.file.Seek(lfile.lastReadPos, 0)
			return -1, nil, err
		}
	}
	lfile.lastReadPos += int64(4)
	size := int(binary.BigEndian.Uint32(sizeBuf))
	data := make([]byte, size)
	n, err = lfile.file.Read(data)
	if err != nil || n != size {
		if errors.Is(err, io.EOF) {
			// we have reached the end of the file
			return -1, nil, err
		} else {
			// file broken (I/O error while writing log at last launch)
			// set lastReadPos to the next log writing point
			lfile.file.Seek(lfile.lastReadPos-4, 0)
			return -1, nil, err
		}
	}
	lfile.lastReadPos += int64(size)
	return n, data, nil
}
