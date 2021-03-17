package main

import (
	"fmt"
	"strings"
	"time"

	libvirt "github.com/libvirt/libvirt-go"
)

func NodeJoin(masterName string, workerName string, masterPass string, workerPass string) (string, error) {
	joinCmd, err := getJoinCmd(masterName, masterPass)
	fmt.Println(joinCmd)
	hostName, retStr, err := joinWorkerNode(workerName, workerPass, joinCmd)
	fmt.Println(hostName)
	fmt.Println(retStr)

	if retStr != "joined" {
		return "", err
	}

	return retStr, nil
}

func joinWorkerNode(workerName string, workerPass string, joinCmd string) (string, string, error) {
	uri := "qemu:///system"
	conn, err := libvirt.NewConnect(uri)
	if err != nil {
		fmt.Println("failed to connect to qemu")
		fmt.Println(err)
		return "", "", err
	}
	defer conn.Close()
	domain, err := conn.LookupDomainByName(workerName)
	if err != nil {
		fmt.Println("domain search fail", err.Error())
		return "", "", err
	}
	defer domain.Free()
	vmState, _, _ := domain.GetState()
	for i := 1; i <= 30; i++ {
		if vmState == libvirt.DOMAIN_RUNNING {
			fmt.Println("workervmState:", vmState)
			break
		}
		time.Sleep(2 * time.Second)
	}

	conn2, err := domain.DomainGetConnect()

	stream, err := conn2.NewStream(0)
	if err != nil {
		fmt.Println(err)
		return "", "", err
	}

	if err := domain.OpenConsole("", stream, 1); err != nil {
		fmt.Println(err)
		return "", "", err
	}

	_, err = stream.Send([]byte("\n"))
	if err != nil {
		fmt.Println(err.Error())
		return "", "", err
	}

	var quit bool
	var checklogin bool
	var str string

	for !quit {
		buf := make([]byte, 1024)
		got, err := stream.Recv(buf)
		if err != nil {
			fmt.Println(err.Error())
			return "", "", err
		}

		if got == 0 {
			break
		}
		str += strings.Trim(string(buf), "\x00")

		// fmt.Println(got, strings.Trim(string(buf), "\x00"))
		// fmt.Println(str)

		var subStr1 string
		var subStr2 string
		if len(str) > 6 {
			subStr1 = str[len(str)-7:]
			subStr2 = str[len(str)-4:]
		}
		// fmt.Println(subStr1)
		// fmt.Println(subStr2)

		if strings.TrimSpace(subStr2) == ":~#" {
			quit = true
		}

		if strings.TrimSpace(subStr1) == "login:" {
			quit = true
			checklogin = true
			fmt.Println("====login====")
		}
	}

	str = ""
	if checklogin == true {
		_, err = stream.Send([]byte("root\n"))
		if err != nil {
			fmt.Println(err.Error())
			return "", "", err
		}
		quit = false

		for !quit {
			buf := make([]byte, 1024)
			got, err := stream.Recv(buf)
			if err != nil {
				fmt.Println(err.Error())
				return "", "", err
			}

			if got == 0 {
				break
			}
			str += strings.Trim(string(buf), "\x00")

			// fmt.Println(str)

			var subStr1 string
			if len(str) > 8 {
				subStr1 = str[len(str)-10:]
			}
			// fmt.Println(subStr1)
			if strings.TrimSpace(subStr1) == "Password:" {
				quit = true
			}
		}
		_, err = stream.Send([]byte(workerPass + "\n"))
		if err != nil {
			fmt.Println(err.Error())
			return "", "", err
		}
		quit = false
		str = ""
		for !quit {
			buf := make([]byte, 1024)
			got, err := stream.Recv(buf)
			if err != nil {
				fmt.Println(err.Error())
				return "", "", err
			}

			if got == 0 {
				break
			}
			str += strings.Trim(string(buf), "\x00")
			fmt.Println(str)

			var subStr1 string
			if len(str) > 6 {
				subStr1 = str[len(str)-3:]
			}
			// fmt.Println(subStr1)
			if strings.TrimSpace(subStr1) == "~#" {
				quit = true
			}
		}
	}
	fmt.Println("1. worker login success")

	cmds := "hostnamectl set-hostname " + workerName
	_, err = stream.Send([]byte(cmds + "\n"))
	if err != nil {
		fmt.Println(err.Error())
		return "", "", err
	}
	quit = false
	str = ""
	for !quit {
		buf := make([]byte, 1024)
		got, err := stream.Recv(buf)
		if err != nil {
			fmt.Println(err.Error())
			return "", "", err
		}

		if got == 0 {
			break
		}
		str += strings.Trim(string(buf), "\x00")
		// fmt.Println(str)

		var subStr1 string
		if len(str) > 6 {
			subStr1 = str[len(str)-3:]
		}
		// fmt.Println(subStr1)
		if strings.TrimSpace(subStr1) == "~#" {
			quit = true
		}

	}
	_, err = stream.Send([]byte("hostname\n"))
	if err != nil {
		fmt.Println(err.Error())
		return "", "", err
	}
	quit = false
	str = ""
	for !quit {
		buf := make([]byte, 1024)
		got, err := stream.Recv(buf)
		if err != nil {
			fmt.Println(err.Error())
			return "", "", err
		}

		if got == 0 {
			break
		}
		str += strings.Trim(string(buf), "\x00")
		// fmt.Println(str)

		var subStr1 string
		if len(str) > 6 {
			subStr1 = str[len(str)-3:]
		}
		// fmt.Println(subStr1)
		if strings.TrimSpace(subStr1) == "~#" {
			quit = true
		}
	}
	var s []string
	var hostName string
	s = strings.Split(str, "\n")
	for _, v := range s {
		// fmt.Println(i, v)
		if strings.Contains(v, workerName) {
			// fmt.Println(v)
			hostName = v
			break
		}
	}
	fmt.Println("2. host name changed")

	_, err = stream.Send([]byte(joinCmd + "\n"))
	if err != nil {
		fmt.Println(err.Error())
		return "", "", err
	}
	quit = false
	str = ""
	for !quit {
		buf := make([]byte, 1024)
		got, err := stream.Recv(buf)
		if err != nil {
			fmt.Println(err.Error())
			return "", "", err
		}

		if got == 0 {
			break
		}
		str += strings.Trim(string(buf), "\x00")
		// fmt.Println(str)

		var subStr1 string
		if len(str) > 6 {
			subStr1 = str[len(str)-3:]
		}
		// fmt.Println(subStr1)
		if strings.TrimSpace(subStr1) == "~#" {
			quit = true
		}
	}

	s = nil
	s = strings.Split(str, "\n")
	var retStr string
	for _, v := range s {
		// fmt.Println(i, v)
		if strings.Contains(v, "This node has joined the cluster:") {
			// fmt.Println(v)
			retStr = "joined"
			break
		}
	}
	fmt.Println("3. node joined")
	return hostName, retStr, nil
}

func getJoinCmd(masterName string, masterPass string) (string, error) {

	uri := "qemu:///system"
	conn, err := libvirt.NewConnect(uri)
	if err != nil {
		fmt.Println("failed to connect to qemu")
		fmt.Println(err)
		return "", err
	}
	defer conn.Close()
	domain, err := conn.LookupDomainByName(masterName)
	if err != nil {
		fmt.Println("domain search fail", err.Error())
		return "", err
	}
	defer domain.Free()

	conn2, err := domain.DomainGetConnect()

	stream, err := conn2.NewStream(0)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	if err := domain.OpenConsole("", stream, 1); err != nil {
		fmt.Println(err)
		return "", err
	}

	_, err = stream.Send([]byte("\n"))
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	var quit bool
	var checklogin bool
	var str string

	for !quit {
		buf := make([]byte, 1024)
		got, err := stream.Recv(buf)
		if err != nil {
			fmt.Println(err.Error())
			return "", err
		}

		if got == 0 {
			break
		}
		str += strings.Trim(string(buf), "\x00")

		// fmt.Println(got, strings.Trim(string(buf), "\x00"))
		// fmt.Println(str)

		var subStr1 string
		var subStr2 string
		if len(str) > 6 {
			subStr1 = str[len(str)-7:]
			subStr2 = str[len(str)-4:]
		}
		// fmt.Println(subStr1)
		// fmt.Println(subStr2)

		if strings.TrimSpace(subStr2) == ":~#" {
			quit = true
		}

		if strings.TrimSpace(subStr1) == "login:" {
			quit = true
			checklogin = true
			fmt.Println("====login====")
		}
	}
	str = ""
	if checklogin == true {
		_, err = stream.Send([]byte("root\n"))
		if err != nil {
			fmt.Println(err.Error())
			return "", err
		}
		quit = false

		for !quit {
			buf := make([]byte, 1024)
			got, err := stream.Recv(buf)
			if err != nil {
				fmt.Println(err.Error())
			}

			if got == 0 {
				break
			}
			str += strings.Trim(string(buf), "\x00")

			// fmt.Println(str)

			var subStr1 string
			if len(str) > 8 {
				subStr1 = str[len(str)-10:]
			}
			// fmt.Println(subStr1)
			if strings.TrimSpace(subStr1) == "Password:" {
				quit = true
			}
		}

		_, err = stream.Send([]byte(masterPass + "\n"))
		if err != nil {
			fmt.Println(err.Error())
			return "", err
		}
		quit = false
		str = ""
		for !quit {
			buf := make([]byte, 1024)
			got, err := stream.Recv(buf)
			if err != nil {
				fmt.Println(err.Error())
				return "", err
			}

			if got == 0 {
				break
			}
			str += strings.Trim(string(buf), "\x00")
			// fmt.Println(str)

			var subStr1 string
			if len(str) > 6 {
				subStr1 = str[len(str)-3:]
			}
			// fmt.Println(subStr1)
			if strings.TrimSpace(subStr1) == "~#" {
				quit = true
			}
		}
	}
	_, err = stream.Send([]byte("kubeadm token create --print-join-command\n"))
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	quit = false
	str = ""
	for !quit {
		buf := make([]byte, 1024)
		got, err := stream.Recv(buf)
		if err != nil {
			fmt.Println(err.Error())
			return "", err
		}

		if got == 0 {
			break
		}
		str += strings.Trim(string(buf), "\x00")
		// fmt.Println(str)

		var subStr1 string
		if len(str) > 6 {
			subStr1 = str[len(str)-3:]
		}
		// fmt.Println(subStr1)
		if strings.TrimSpace(subStr1) == "~#" {
			quit = true
		}
	}

	var s []string
	var joinCmd string
	if strings.Contains(str, "kubeadm join") {
		s = strings.Split(str, "\n")
	}
	for _, v := range s {
		// fmt.Println(i, v)
		if strings.Contains(v, "kubeadm join") {
			joinCmd = v
			break
		}
	}
	return joinCmd, nil
}
