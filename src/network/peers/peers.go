package peers

import (
	"fmt"
	"net"
	"sort"
	"time"

	"../../def"
	"../../ordermanager"
	"../conn"
)

type PeerUpdate struct {
	Peers []string
	New   string
	Lost  []string
}

const interval = 15 * time.Millisecond
const timeout = 50 * time.Millisecond

func Transmitter(port int, id string, transmitEnable <-chan bool) {

	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))

	enable := true
	for {
		select {
		case enable = <-transmitEnable:
		case <-time.After(interval):
		}
		if enable {
			conn.WriteTo([]byte(id), addr)
		}
	}
}

func Receiver(port int, peerUpdateCh chan<- PeerUpdate) {

	var buf [1024]byte
	var p PeerUpdate
	lastSeen := make(map[string]time.Time)

	conn := conn.DialBroadcastUDP(port)

	for {
		updated := false

		conn.SetReadDeadline(time.Now().Add(interval))
		n, _, _ := conn.ReadFrom(buf[0:])

		id := string(buf[:n])

		// Adding new connection
		p.New = ""
		if id != "" {
			if _, idExists := lastSeen[id]; !idExists {
				p.New = id
				updated = true
			}

			lastSeen[id] = time.Now()
		}

		// Removing dead connection
		p.Lost = make([]string, 0)
		for k, v := range lastSeen {
			if time.Now().Sub(v) > timeout {
				updated = true
				p.Lost = append(p.Lost, k)
				delete(lastSeen, k)
			}
		}

		// Sending update
		if updated {
			p.Peers = make([]string, 0, len(lastSeen))

			for k, _ := range lastSeen {
				p.Peers = append(p.Peers, k)
			}

			sort.Strings(p.Peers)
			sort.Strings(p.Lost)
			peerUpdateCh <- p
		}
	}
}

func PeerWatch(msg_deadElev chan<- def.MapMessage) {
	transmitEnable := make(chan bool, 100)
	peerUpdateCh := make(chan PeerUpdate, 100)

	var sendID string
	var ID int

	switch def.LOCAL_ID {
	// To have a more unique ID to send and recieve, we chose to make this switch.
	// To allow for more elevators, make new cases with unique strings as ID's.
	case 0:
		sendID = "sendIDiszero"
	case 1:
		sendID = "sendIDisone"
	case 2:
		sendID = "sendIDistwo"
	}

	go Transmitter(def.SEND_ID_PORT, sendID, transmitEnable)
	go PollNetwork(peerUpdateCh)

	var currentMap ordermanager.ElevatorMap
	var send bool

	for {
		send = false
		select {
		case msg := <-peerUpdateCh:
			currentMap = ordermanager.GetElevMap()
			if msg.New != "" {
				switch msg.New {
				case "sendIDiszero":
					ID = 0
					send = true
				case "sendIDisone":
					ID = 1
					send = true
				case "sendIDistwo":
					ID = 2
					send = true

				default:
					ID = -1
					send = false
				}

				if send && ID != -1 {
					currentMap[ID].State = def.S_Idle

					sendMsg := def.MakeMapMessage(currentMap, "New elevator")
					msg_deadElev <- sendMsg
				}

			} else if len(msg.Lost) > 0 {
				if msg.Lost[0] != "" {
					switch msg.Lost[0] {
					case "sendIDiszero":
						ID = 0
						send = true
					case "sendIDisone":
						ID = 1
						send = true
					case "sendIDistwo":
						ID = 2
						send = true

					default:
						ID = -1
						send = false
					}
					if send {
						currentMap[ID].State = def.S_Dead

						sendMsg := def.MakeMapMessage(currentMap, "Dead elevator")
						msg_deadElev <- sendMsg
					}
				}
			}

		}
	}
}

func PollNetwork(peerUpdateCh chan<- PeerUpdate) {
	poll_chn := make(chan PeerUpdate, 100)

	for port := 20010; port < 20100; port++ {
		if port != def.SEND_ID_PORT {
			go Receiver(port, poll_chn)
		}
	}

	for {
		select {
		case msg_fromNet := <-poll_chn:
			peerUpdateCh <- msg_fromNet
		}
	}
}
