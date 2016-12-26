package main

import (
	"github.com/tarm/serial"
	"gopkg.in/yaml.v2"
	"log"
	"net/smtp"
	"time"
	"io/ioutil"
	"os"
)

var pingTimer *time.Timer
var config *Config

type Config struct {
	Serial struct {
		device string
		baud   int
	}
	SMTP struct {
		host     string
		port     int
		user     string
		password string
	}
}

func main() {
	config = parseConfig(os.Args[1:][0])
	readSerial()
}

func parseConfig(file string)  (*Config) {

	data, err := ioutil.ReadFile(file)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	t := Config{}
	err = yaml.Unmarshal([]byte(data), &t)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return &t
}

func readSerial() {

	for {
		port := connectToSerial()

		for {
			command, err := readCommandFromSerial(port)
			if err != nil {
				log.Println(err.Error())
				break
			} else {
				handleCommand(command)
			}
		}

		port.Close()
	}
}

func connectToSerial() *serial.Port {
	c := &serial.Config{Name: config.Serial.device, Baud: config.Serial.baud}

	for {
		s, err := serial.OpenPort(c)

		if err != nil {
			log.Println(err.Error())
			time.Sleep(time.Duration(1) * time.Second)
			continue
		} else {
			log.Println("Connected")
			setPingTimer()
			s.Flush()
			return s
		}
	}
}

func readCommandFromSerial(s *serial.Port) (string, error) {
	command := make([]byte, 0)

	for readLength := 0; readLength < 5; {

		buf := make([]byte, 1)
		n, err := s.Read(buf)
		if err != nil {
			return string(command[:]), err
		}
		readLength += n
		command = append(command, buf[:n]...)
	}

	return string(command[:3]), nil
}

//TODO: guaranteed delivery
func sendMail(command string) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	c, err := smtp.Dial("smtp.gmail.com:587")

	if err != nil {
		log.Fatal(err)
	}

	c.Verify("aleksandr.zhuikov@gmail.com")

	// Set up authentication information.
	auth := smtp.PlainAuth("", "aleksandr.zhuikov@gmail.com", "timwaxgbyfucdtsg", "smtp.gmail.com")

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	to := []string{"aleksandr.zhuikov@gmail.com"}

	msg := []byte("Subject: [TEPEMOK] " + command + "\r\n" +
		"\r\n" +
		command)
	err = smtp.SendMail("smtp.gmail.com:587", auth, "aleksandr.zhuikov@gmail.com", to, []byte(msg))
}

func handleCommand(command string) {
	log.Printf("%q", command)

	switch []rune(command)[0] {
	case 'H':
		resetPingTime()
		log.Printf("ping")
	case 'N':
		sendMail("alarm raised " + command)
		log.Printf("alarm raised " + command)
	case 'Y':
		sendMail("alarm resolved " + command)
		log.Printf("alarm resolved " + command)
	default:
		log.Printf("Unknown command " + command)
	}
}

func setPingTimer() {
	pingTimer = time.NewTimer(time.Second * 5)
	go func() {
		<-pingTimer.C
		sendMail("no ping for too long")
		log.Printf("No ping for too long")
	}()
}

func resetPingTime() {
	pingTimer.Reset(time.Second * 5)
}
