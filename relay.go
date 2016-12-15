package main

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"time"
)

//const hostname = "localhost"
//const port = "8876"
const hostname = "luna.red"
const port = "44444"

//const unixPath = "/var/www/bot/irc/sock"
const password = "bokunopico911"

const nickname = "SystemD"

/*
type Config struct {
    hostname string
    port string
    password
}

type misc struct {
    Nickname string
}

var config Config
*/
//Creates the inital connection to the irc server
func connect() net.Conn {
	ircConnection, err := net.Dial("tcp", hostname+":"+port)

	if err != nil {
		panic(err)
	}

	sendCommand(ircConnection, "PASS", password)
	sendCommand(ircConnection, "USER", fmt.Sprintf("%s 8 * :%s", nickname, nickname))
	sendCommand(ircConnection, "NICK", nickname)

	return ircConnection
}

func sendCommand(conn net.Conn, command string, text string) {
	fmt.Fprintf(conn, "%s %s\n", command, text)
}

func getListener() net.Listener {

	sockListener, err := net.Listen("tcp", "localhost:8765")

	if err != nil {
		panic(err)
	}

	return sockListener
}

func runBot() {
	_, err := exec.Command("../irc/irc").Output()
	if err != nil {
		panic(err)
	}
}

func reloadBot() {
	_, err := exec.Command("../irc/reloadBot.fsh").Output()
	if err != nil {
		panic(err)
	}
}

func ircListener(ircConn net.Conn, botConn *net.Conn) {
	scanner := bufio.NewScanner(ircConn)
	for scanner.Scan() {
		msg := scanner.Text()
		fmt.Printf("FROM IRC: %s \n", msg)
		rePing := regexp.MustCompile(`PING :(.+)$`)
		rePriv := regexp.MustCompile(`:([^!]+)!([^@]+)@([^ ]+) PRIVMSG ([^ ]+) :(.+)$`)

		find := func(reg *regexp.Regexp, msg string) []string {
			return reg.FindStringSubmatch(msg)
		}

		fmt.Printf("RECIEVED: %s\n", msg)
		if rePing.MatchString(msg) {
			go pingHandler(ircConn, find(rePing, msg))
		} else if rePriv.MatchString(msg) {
			command := find(rePriv, msg)[5]
			if command == "!reload" {
				go reloadBot()
			}
		} else {
			if botConn != nil && *botConn != nil {
				fmt.Println(botConn)
				fmt.Println(*botConn)
				fmt.Fprintf(*botConn, "%s\r\n", msg)
			}
		}
	}
}

func botListener(ircConn net.Conn, botConn *net.Conn) {
	for {
		if botConn == nil || *botConn == nil {
			fmt.Println("Tried making scanner, botConn is: ", botConn)
			time.Sleep(2 * time.Second)
			continue
		}
		scanner := bufio.NewScanner(*botConn)
		for scanner.Scan() {
			msg := scanner.Text()
			fmt.Printf("FROM BOT: %s \n", msg)
			fmt.Fprintf(ircConn, "%s\r\n", msg)
		}
	}
}

func multiplexer(ircConn net.Conn, botConn *net.Conn) {
	go ircListener(ircConn, botConn)
	botListener(ircConn, botConn)
}

func main() {
	//err := ini.MapTo(config, "relay.ini`")

	ircConn := connect()
	fmt.Println("Got irc connection")

	unixListener := getListener()
	fmt.Println("Got listener!")

	defer ircConn.Close()
	defer unixListener.Close()

	var botConn net.Conn

	//go multiplexer(ircConn, botConn)
	go ircListener(ircConn, &botConn)
	go botListener(ircConn, &botConn)

	var err error
	//go startListener(botConn, unixListener)
	for {
		botConn, err = unixListener.Accept()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Accepted one connection")
	}
	//startListener(botConn, unixListener)
	//    go runBot()
}

func startListener(botConn *net.Conn, listener net.Listener) {
	var err error

	for {
		*botConn, err = listener.Accept()
		if err != nil {
			panic(err)
		}
		fmt.Println("Accepted one connection")
	}
}
