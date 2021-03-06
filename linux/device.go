package linux

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"syscall"
	"unsafe"

	"github.com/develersrl/gatt/linux/gioctl"
	"github.com/develersrl/gatt/linux/socket"
)

type device struct {
	fd   int
	dev  int
	name string
	rmu  *sync.Mutex
	wmu  *sync.Mutex
}

func newDevice(n int, chk bool) (*device, error) {
	fmt.Printf("LA MARZOCCO - socket.Socket(socket.AF_BLUETOOTH, syscall.SOCK_RAW, socket.BTPROTO_HCI)\n")
	fd, err := socket.Socket(socket.AF_BLUETOOTH, syscall.SOCK_RAW, socket.BTPROTO_HCI)
	fmt.Printf("LA MARZOCCO - fd=%v, err=%v\n", fd, err)

	if err != nil {
		return nil, err
	}
	if n != -1 {
		return newSocket(fd, n, chk)
	}

	req := devListRequest{devNum: hciMaxDevices}
	fmt.Printf("LA MARZOCCO - gioctl.Ioctl(uintptr(fd)=%v, hciGetDeviceList, uintptr(unsafe.Pointer(&req)))\n", uintptr(fd))
	err = gioctl.Ioctl(uintptr(fd), hciGetDeviceList, uintptr(unsafe.Pointer(&req)))
	fmt.Printf("LA MARZOCCO - err=%v\n", err)
	if err != nil {
		return nil, err
	}
	fmt.Printf("LA MARZOCCO - int(req.devNum)=%v\n", int(req.devNum))
	for i := 0; i < int(req.devNum); i++ {
		d, err := newSocket(fd, i, chk)
		if err == nil {
			log.Printf("dev: %s opened", d.name)
			return d, err
		}
	}
	return nil, errors.New("no supported devices available")
}

func newSocket(fd, n int, chk bool) (*device, error) {
	i := hciDevInfo{id: uint16(n)}

	fmt.Printf("LA MARZOCCO - gioctl.Ioctl(uintptr(fd)=%v, hciGetDeviceInfo, uintptr(unsafe.Pointer(&i)))\n", uintptr(fd))
	err := gioctl.Ioctl(uintptr(fd), hciGetDeviceInfo, uintptr(unsafe.Pointer(&i)))
	fmt.Printf("LA MARZOCCO - err=%v\n", err)

	if err != nil {
		return nil, err
	}
	name := string(i.name[:])
	// Check the feature list returned feature list.
	if chk && i.features[4]&0x40 == 0 {
		err := errors.New("does not support LE")
		log.Printf("dev: %s %s", name, err)
		return nil, err
	}
	log.Printf("dev: %s up", name)
	fmt.Printf("LA MARZOCCO - gioctl.Ioctl(uintptr(fd)=%v, hciUpDevice, uintptr(n)=%v)\n", uintptr(fd), uintptr(n))
	err = gioctl.Ioctl(uintptr(fd), hciUpDevice, uintptr(n))
	fmt.Printf("LA MARZOCCO - err=%v\n", err)
	if err != nil {
		if err != syscall.EALREADY {
			return nil, err
		}
		log.Printf("dev: %s reset", name)
		fmt.Printf("LA MARZOCCO - gioctl.Ioctl(uintptr(fd)=%v, hciResetDevice, uintptr(n)=%v)\n", uintptr(fd), uintptr(n))
		err = gioctl.Ioctl(uintptr(fd), hciResetDevice, uintptr(n))
		fmt.Printf("LA MARZOCCO - err=%v\n", err)
		if err != nil {
			return nil, err
		}
	}
	log.Printf("dev: %s down", name)
	fmt.Printf("LA MARZOCCO - gioctl.Ioctl(uintptr(fd)=%v, hciDownDevice, uintptr(n)=%v)\n", uintptr(fd), uintptr(n))
	err = gioctl.Ioctl(uintptr(fd), hciDownDevice, uintptr(n))
	fmt.Printf("LA MARZOCCO - err=%v\n", err)
	if err != nil {
		return nil, err
	}

	// Attempt to use the linux 3.14 feature, if this fails with EINVAL fall back to raw access
	// on older kernels.
	sa := socket.SockaddrHCI{Dev: n, Channel: socket.HCI_CHANNEL_USER}
	fmt.Printf("sa := socket.SockaddrHCI{Dev: %v, Channel: socket.HCI_CHANNEL_USER}", n)
	fmt.Printf("LA MARZOCCO - socket.Bind(fd=%v, &sa)\n", fd)
	err = socket.Bind(fd, &sa)
	fmt.Printf("LA MARZOCCO - err=%v\n", err)
	if err != nil {
		if err != syscall.EINVAL {
			return nil, err
		}
		log.Printf("dev: %s can't bind to hci user channel, err: %s.", name, err)
		sa := socket.SockaddrHCI{Dev: n, Channel: socket.HCI_CHANNEL_RAW}
		fmt.Printf("sa := socket.SockaddrHCI{Dev: %v, Channel: socket.HCI_CHANNEL_RAW}", n)
		fmt.Printf("LA MARZOCCO - socket.Bind(fd=%v, &sa)\n", fd)
		err = socket.Bind(fd, &sa)
		fmt.Printf("LA MARZOCCO - err=%v\n", err)
		if err != nil {
			log.Printf("dev: %s can't bind to hci raw channel, err: %s.", name, err)
			return nil, err
		}
	}
	return &device{
		fd:   fd,
		dev:  n,
		name: name,
		rmu:  &sync.Mutex{},
		wmu:  &sync.Mutex{},
	}, nil
}

func (d device) Read(b []byte) (int, error) {
	d.rmu.Lock()
	defer d.rmu.Unlock()
	fmt.Printf("syscall.Read(d.fd=%v, b=%v)\n", d.fd, b)
	n, err := syscall.Read(d.fd, b)
	fmt.Printf("n=%v err=%v\n", n, err)
	return n, err
}

func (d device) Write(b []byte) (int, error) {
	d.wmu.Lock()
	defer d.wmu.Unlock()
	fmt.Printf("syscall.Write(d.fd=%v, b=%v)\n", d.fd, b)
	n, err := syscall.Write(d.fd, b)
	fmt.Printf("n=%v err=%v\n", n, err)
	return n, err
}

func (d device) Close() error {
	fmt.Printf("syscall.Close(d.fd=%v)\n", d.fd)
	err := syscall.Close(d.fd)
	fmt.Printf("err=%v\n", err)
	return err
}
