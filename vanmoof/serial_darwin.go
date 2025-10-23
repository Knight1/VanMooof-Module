//go:build darwin

package vanmoof

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// DarwinSerialPort represents a macOS serial port
type DarwinSerialPort struct {
	file *os.File
}

// openSerial opens a serial port on macOS
func openSerial(port string, baudRate uint32) (SerialPort, error) {
	file, err := os.OpenFile(port, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open port %s: %v", port, err)
	}

	// Configure serial port using termios
	fd := int(file.Fd())

	// Get current termios settings
	var termios unix.Termios
	if err := getTermios(fd, &termios); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to get termios: %v", err)
	}

	// Configure for raw mode
	termios.Cflag &^= unix.PARENB // No parity
	termios.Cflag &^= unix.CSTOPB // 1 stop bit
	termios.Cflag &^= unix.CSIZE  // Clear size bits
	termios.Cflag |= unix.CS8     // 8 data bits
	termios.Cflag |= unix.CREAD   // Enable receiver
	termios.Cflag |= unix.CLOCAL  // Ignore modem control lines

	// Disable canonical mode and echo
	termios.Lflag &^= unix.ICANON
	termios.Lflag &^= unix.ECHO
	termios.Lflag &^= unix.ECHOE
	termios.Lflag &^= unix.ISIG

	// Disable input processing
	termios.Iflag &^= unix.IXON
	termios.Iflag &^= unix.IXOFF
	termios.Iflag &^= unix.IXANY
	termios.Iflag &^= unix.INLCR
	termios.Iflag &^= unix.ICRNL

	// Disable output processing
	termios.Oflag &^= unix.OPOST

	// Set baud rate to 115200 (always used)
	termios.Cflag |= unix.B115200

	// Set timeouts
	termios.Cc[unix.VMIN] = 0   // Minimum characters to read
	termios.Cc[unix.VTIME] = 10 // Timeout in deciseconds

	// Apply settings
	if err := setTermios(fd, &termios); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to set termios: %v", err)
	}

	return &DarwinSerialPort{file: file}, nil
}

// Read implements io.Reader
func (sp *DarwinSerialPort) Read(p []byte) (n int, err error) {
	return sp.file.Read(p)
}

// Write implements io.Writer
func (sp *DarwinSerialPort) Write(p []byte) (n int, err error) {
	return sp.file.Write(p)
}

// Close closes the serial port
func (sp *DarwinSerialPort) Close() error {
	return sp.file.Close()
}

// getTermios gets termios structure (macOS)
func getTermios(fd int, termios *unix.Termios) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), unix.TIOCGETA, uintptr(unsafe.Pointer(termios)))
	if errno != 0 {
		return errno
	}
	return nil
}

// setTermios sets termios structure (macOS)
func setTermios(fd int, termios *unix.Termios) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), unix.TIOCSETA, uintptr(unsafe.Pointer(termios)))
	if errno != 0 {
		return errno
	}
	return nil
}
