package main

import (
	"github.com/tarm/serial"
	"gopkg.in/yaml.v2"
	"log"
	"net/smtp"
	"time"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var pingTimer *time.Timer
var config *Config

type Config struct {
	Serial struct {
		Device string
		Baud   int
	}
	SMTP struct {
		Host     string
		Port     int
		User     string
		Password string
		From string
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
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

	log.Println("Parsed config from " + file)

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

	log.Println("Will use serial port " + config.Serial.Device)

	c := &serial.Config{Name: config.Serial.Device, Baud: config.Serial.Baud}

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

	for readLength :=0; !isFullCommand(command); {
		buf := make([]byte, 1)
		n, err := s.Read(buf)
		if err != nil {
			return string(command[:]), err
		}
		readLength += n
		command = append(command, buf[:n]...)
	}

	return string(command), nil
}

func isFullCommand(command []byte) (bool)  {

	if (len(command) < 3) {
		return false
	}

	return strings.HasSuffix(string(command), "\r\n")
}

//TODO: guaranteed delivery
func sendMail(command string) {

	c, err := smtp.Dial(config.SMTP.Host + ":" + strconv.Itoa(config.SMTP.Port))

	if err != nil {
		log.Fatal(err)
	}

	c.Verify(config.SMTP.User)

	// Set up authentication information.
	auth := smtp.PlainAuth("", config.SMTP.User, config.SMTP.Password, config.SMTP.Host)

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	to := []string{config.SMTP.User}

	msg := []byte("Subject: [TEPEMOK] " + command + "\r\n" +
		"\r\n" +
		command)
	err = smtp.SendMail(config.SMTP.Host + ":" + strconv.Itoa(config.SMTP.Port), auth, config.SMTP.From, to, []byte(msg))
}

func handleCommand(command string) {
	log.Printf("%q", command)

	switch []rune(command)[0] {
	case 'H':
		resetPingTime()
		log.Printf("ping")
	case 'A':
		sendMail("alarm raised " + command)
		log.Printf("alarm raised " + command)
	case 'Y':
		sendMail("alarm resolved " + command)
		log.Printf("alarm resolved " + command)
	case 'C':
		sendMail("wire cut " + command)
		log.Printf("wire cut " + command)
	case 'S':
		sendMail("short circuit " + command)
		log.Printf("short circuit " + command)
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
