//go:build windows

package vanmoof

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// WindowsSerialPort represents a Windows serial port
type WindowsSerialPort struct {
	handle windows.Handle
}

// openSerial opens a serial port on Windows
func openSerial(port string, baudRate uint32) (SerialPort, error) {
	portName := fmt.Sprintf("\\\\.\\%s", port)
	portNameUTF16, err := windows.UTF16PtrFromString(portName)
	if err != nil {
		return nil, err
	}

	handle, err := windows.CreateFile(
		portNameUTF16,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		0,
		nil,
		windows.OPEN_EXISTING,
		0,
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open port %s: %v", port, err)
	}

	// Configure serial port
	var dcb windows.DCB
	dcb.DCBlength = uint32(unsafe.Sizeof(dcb))
	err = windows.GetCommState(handle, &dcb)
	if err != nil {
		if closeErr := windows.CloseHandle(handle); closeErr != nil {
			return nil, fmt.Errorf("failed to get comm state: %v (close error: %v)", err, closeErr)
		}
		return nil, fmt.Errorf("failed to get comm state: %v", err)
	}

	dcb.BaudRate = baudRate
	dcb.ByteSize = 8
	dcb.Parity = 0   // No parity
	dcb.StopBits = 0 // 1 stop bit

	err = windows.SetCommState(handle, &dcb)
	if err != nil {
		if closeErr := windows.CloseHandle(handle); closeErr != nil {
			return nil, fmt.Errorf("failed to set comm state: %v (close error: %v)", err, closeErr)
		}
		return nil, fmt.Errorf("failed to set comm state: %v", err)
	}

	// Set timeouts
	timeouts := windows.CommTimeouts{
		ReadIntervalTimeout:         50,
		ReadTotalTimeoutMultiplier:  10,
		ReadTotalTimeoutConstant:    1000,
		WriteTotalTimeoutMultiplier: 10,
		WriteTotalTimeoutConstant:   1000,
	}
	err = windows.SetCommTimeouts(handle, &timeouts)
	if err != nil {
		if closeErr := windows.CloseHandle(handle); closeErr != nil {
			return nil, fmt.Errorf("failed to set timeouts: %v (close error: %v)", err, closeErr)
		}
		return nil, fmt.Errorf("failed to set timeouts: %v", err)
	}

	return &WindowsSerialPort{handle: handle}, nil
}

// Read implements io.Reader
func (sp *WindowsSerialPort) Read(p []byte) (n int, err error) {
	var bytesRead uint32
	err = windows.ReadFile(sp.handle, p, &bytesRead, nil)
	return int(bytesRead), err
}

// Write implements io.Writer
func (sp *WindowsSerialPort) Write(p []byte) (n int, err error) {
	var bytesWritten uint32
	err = windows.WriteFile(sp.handle, p, &bytesWritten, nil)
	return int(bytesWritten), err
}

// Close closes the serial port
func (sp *WindowsSerialPort) Close() error {
	return windows.CloseHandle(sp.handle)
}
