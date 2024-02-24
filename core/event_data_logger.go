package core

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"os"
	"sync"
)

type EventDataLogger struct {
	logFile     *os.File
	logFileMtx  *sync.Mutex
	curSize     int64
	lastReadPos int64
	// when recovery from log file, set this to false
	IsLoggingActive bool
}

func NewEventDataLogger(logFilename string) *EventDataLogger {
	file, err := os.OpenFile(logFilename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalln("can't open db file: " + logFilename)
		return nil
	}
	file.Seek(io.SeekStart, 0)
	return &EventDataLogger{
		logFile:         file,
		logFileMtx:      &sync.Mutex{},
		curSize:         -1,
		lastReadPos:     0,
		IsLoggingActive: true,
	}
}

func (l *EventDataLogger) GetLogfileSize() int64 {
	fi, err := l.logFile.Stat()
	if err != nil {
		panic(err)
	}
	return fi.Size()
}

func (l *EventDataLogger) WriteLog(data []byte) error {
	if !l.IsLoggingActive {
		return nil
	}

	l.logFileMtx.Lock()
	defer l.logFileMtx.Unlock()
	if l.curSize < 0 {
		fileInfo, err := l.logFile.Stat()
		if err != nil {
			return err
		}
		l.curSize = fileInfo.Size()
		l.logFile.Seek(l.curSize, 0)
	}
	sizeBuf := make([]byte, 0)
	sizeBuf = binary.LittleEndian.AppendUint32(sizeBuf, uint32(len(data)))
	// each log entry is prefixed with a 4-byte size
	n, err := l.logFile.Write(sizeBuf)
	if err != nil || n != 4 {
		panic(err)
	}
	n, err = l.logFile.Write(data)
	if err != nil || n != len(data) {
		panic(err)
	}
	err = l.logFile.Sync()
	if err != nil {
		panic(err)
	}
	l.curSize += int64(n + 4)

	return err
}

// read a log entry
func (l *EventDataLogger) ReadLog() (int, []byte, error) {
	l.logFileMtx.Lock()
	defer l.logFileMtx.Unlock()
	sizeBuf := make([]byte, 4)
	n, err := l.logFile.Read(sizeBuf)
	if err != nil || n != 4 {
		if errors.Is(err, io.EOF) {
			// we have reached the end of the file
			return -1, nil, err
		} else {
			// file broken (I/O error while writing log at last launch)
			// set lastReadPos to the next log writing point
			l.logFile.Seek(l.lastReadPos, 0)
			return -1, nil, err
		}
	}
	l.lastReadPos += int64(4)
	size := int(binary.LittleEndian.Uint32(sizeBuf))
	data := make([]byte, size)
	n, err = l.logFile.Read(data)
	if err != nil || n != size {
		if errors.Is(err, io.EOF) {
			// we have reached the end of the file
			return -1, nil, err
		} else {
			// file broken (I/O error while writing log at last launch)
			// set lastReadPos to the next log writing point
			l.logFile.Seek(l.lastReadPos-4, 0)
			return -1, nil, err
		}
	}
	l.lastReadPos += int64(size)
	return n, data, nil
}
