package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	//เก็บโฟลเดอร์ที่ผู้ใช้เลือก
	source := ""
	dest := ""
	// อ่านข้อความจาก Terminal ทีละบรรทัด
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("=== Flash Drive Backup ===")
	fmt.Println("/source <path>, /dest <path>")
	fmt.Println("/list source, /list dest, /exit")

	for {
		fmt.Print("> ")
		//ถ้าอ่านข้อมูลไม่ได้จะหยุด
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text()) //ตัดช่องว่างหน้า-ท้าย
		//ถ้าเป็นนค่าว่างจะไม่ทำอะไร
		if input == "" {
			continue
		}

		parts := strings.SplitN(input, " ", 2) //แยกข้อความออกสองส่วน
		command := parts[0]
		argument := ""

		if len(parts) == 2 {
			argument = strings.TrimSpace(parts[1])
			argument = strings.Trim(argument, "\"")
		}
		switch command {
		case "/source", "/dest":
			if argument == "" {
				fmt.Println("กรุณาระบุ path")
				continue
			}
			// แปลงเป็น path เต็มตามระบบ Windows หรือ Ubuntu
			path, err := filepath.Abs(argument)
			if err != nil {
				fmt.Println("Path ไม่ถูกต้อง:", err)
				continue
			}
			_, err = os.ReadDir(path)
			if err != nil {
				fmt.Println("เปิดโฟลเดอร์ไม่ได้:", err)
				continue
			}

			if command == "/source" {
				source = path
				fmt.Println("ตั้งต้นทาง:", source)
			} else {
				dest = path
				fmt.Println("ตั้งปลายทาง:", dest)
			}
		case "/list":
			if argument == "source" {
				listFiles(source)
			} else if argument == "dest" {
				listFiles(dest)
			} else {
				fmt.Println("ใช้ /list source หรือ /list dest")
			}
		case "/exit":
			fmt.Println("จบการทำงาน")
			return
		default:
			fmt.Println("ไม่รู้จักคำสั่ง:", command)
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("อ่านคำสั่งไม่ได้:", err)
		}
	}
}
func listFiles(path string) {
	if path == "" {
		fmt.Println("กรุณาตั้งค่าโฟลเดอร์ก่อน")
		return
	}
	files, err := os.ReadDir(path)
	if err != nil {
		fmt.Println("อ่านโฟลเดอร์ไม่ได้:", err)
		return
	}
	for _, file := range files {
		if !file.IsDir() {
			fmt.Println("-", file.Name())
		}
	}

}
