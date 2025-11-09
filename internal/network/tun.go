package network

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

type TunManager struct {
	device *os.File
	name   string
}

func NewTunManager() *TunManager {
	return &TunManager{}
}

func (tm *TunManager) Create(name string) error {
	fd, err := syscall.Open("/dev/net/tun", syscall.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("failed to open TUN device: %w", err)
	}

	var ifr struct {
		name  [16]byte
		flags uint16
		pad   [22]byte
	}

	copy(ifr.name[:], name)
	ifr.flags = syscall.IFF_TUN | syscall.IFF_NO_PI

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TUNSETIFF, uintptr(unsafe.Pointer(&ifr)))
	if errno != 0 {
		syscall.Close(fd)
		return fmt.Errorf("failed to create TUN interface: %v", errno)
	}

	tm.device = os.NewFile(uintptr(fd), "/dev/net/tun")
	tm.name = name

	if err := tm.configureInterface(); err != nil {
		tm.Close()
		return fmt.Errorf("failed to configure interface: %w", err)
	}

	return nil
}

func (tm *TunManager) configureInterface() error {
	cmd := exec.Command("ip", "link", "set", tm.name, "up")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring interface up: %w", err)
	}

	cmd = exec.Command("ip", "addr", "add", "10.0.0.1/24", "dev", tm.name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set IP address: %w", err)
	}

	return nil
}

func (tm *TunManager) ConfigureClientInterface(clientIP string) error {
	cmd := exec.Command("ip", "link", "set", tm.name, "up")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring interface up: %w", err)
	}

	cmd = exec.Command("ip", "addr", "add", clientIP+"/24", "dev", tm.name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set client IP address: %w", err)
	}

	return nil
}

func (tm *TunManager) ReadPacket() ([]byte, error) {
	if tm.device == nil {
		return nil, fmt.Errorf("TUN interface not created")
	}

	buffer := make([]byte, 1500)
	n, err := tm.device.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read packet: %w", err)
	}

	return buffer[:n], nil
}

func (tm *TunManager) WritePacket(data []byte) error {
	if tm.device == nil {
		return fmt.Errorf("TUN interface not created")
	}

	_, err := tm.device.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write packet: %w", err)
	}

	return nil
}

func (tm *TunManager) Close() error {
	if tm.device == nil {
		return nil
	}

	err := tm.device.Close()
	tm.device = nil
	tm.name = ""

	return err
}

func (tm *TunManager) GetName() string {
	return tm.name
}

func (tm *TunManager) IsCreated() bool {
	return tm.device != nil
}
