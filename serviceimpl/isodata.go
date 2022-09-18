package serviceimpl

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/moov-io/iso8583"
	"github.com/moov-io/iso8583/encoding"
	"github.com/moov-io/iso8583/field"
	"github.com/moov-io/iso8583/padding"
	"github.com/moov-io/iso8583/prefix"
)

type Data struct {
	// MTI string
	Mode string `json:mode`
	PAN  string `json:"pan"`
	PC   int64  `json:"pc"`
	TA   string `json:"ta"`
	SA   string `json:"sa"`
	BA   string `json:"ba"`
	// CAN *Iso43data     `json:"can"`
}

// var addr = flag.String("addr", "192.168.1.44:8500", "http service address")

// type Iso43data struct {
// 	F1 *field.String `json:"f1"`
// 	F2 *field.String `json:"f2"`
// 	F3 *field.String `json:"f3"`
// 	F4 *field.String `json:"f4"`
// }

type ClientManager struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

type Client struct {
	socket net.Conn
	data   chan []byte
}

func (manager *ClientManager) start() {
	for {
		select {
		case connection := <-manager.register:
			manager.clients[connection] = true
			fmt.Println("Added new connection!")
		case connection := <-manager.unregister:
			if _, ok := manager.clients[connection]; ok {
				close(connection.data)
				delete(manager.clients, connection)
				fmt.Println("A connection has terminated!")
			}
		case message := <-manager.broadcast:
			for connection := range manager.clients {
				select {
				case connection.data <- message:
				default:
					close(connection.data)
					delete(manager.clients, connection)
				}
			}
		}
	}
}

func (manager *ClientManager) receive(client *Client) {
	for {
		message := make([]byte, 4096)
		length, err := client.socket.Read(message)
		if err != nil {
			manager.unregister <- client
			client.socket.Close()
			break
		}
		if length > 0 {
			fmt.Println("RECEIVED: " + string(message))
			manager.broadcast <- message
		}
	}
}

func (client *Client) receive() {
	for {
		message := make([]byte, 4096)
		length, err := client.socket.Read(message)
		if err != nil {
			client.socket.Close()
			break
		}
		if length > 0 {
			fmt.Println("RECEIVED: " + string(message))
		}
	}
}

func (client *Client) send() {
	for {
		messages := make([]byte, 4096)
		length, err := client.socket.Write(messages)
		if err != nil {
			client.socket.Close()
			break
		}
		if length > 0 {
			fmt.Println("Sending: " + string(messages))
		}
	}
}

func (manager *ClientManager) send(client *Client) {
	defer client.socket.Close()
	for {
		select {
		case message, ok := <-client.data:
			if !ok {
				return
			}
			client.socket.Write(message)
		}
	}
}

func startServerMode() {
	fmt.Println("Starting server...")
	listener, error := net.Listen("tcp", ":8500")
	if error != nil {
		fmt.Println(error)
	}
	fmt.Println("Connection succesful")
	manager := ClientManager{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	go manager.start()
	fmt.Println("Connection succesful manager")
	for {
		connection, _ := listener.Accept()
		if error != nil {
			fmt.Println(error)
		}
		client := &Client{socket: connection, data: make(chan []byte)}
		manager.register <- client
		go manager.receive(client)
		go manager.send(client)
		fmt.Println("Connection succesful server")
	}
}

func startClientMode(rawMessage []byte) {
	fmt.Println("Starting client...", string(rawMessage))

	connection, err := net.Dial("tcp", "192.168.1.252:8500")
	if err != nil {
		fmt.Println("Error in client ", err)
	}

	fmt.Println("connected ", connection)
	// defer connection.Close()
	// flag.Parse()
	// log.SetFlags(0)

	// interrupt := make(chan os.Signal, 1)
	// signal.Notify(interrupt, os.Interrupt)

	// //u := url.URL{Scheme: "http", Host: *addr}
	// // log.Printf("connecting to %s", u.String())

	// //websocket.Dialer.NetDial
	// connection, _, err := websocket.DefaultDialer.Dial("192.168.1.44:8500", nil)
	// if err != nil {
	// 	log.Fatal("dial:", err)
	// }
	// // defer connection.Close()

	// // done := make(chan struct{})

	// go func() {
	// 	defer close(done)
	// 	for {
	// 		_, message, err := connection.ReadMessage()
	// 		if err != nil {
	// 			log.Println("read:", err)
	// 			return
	// 		}
	// 		log.Printf("recv: %s", message)
	// 	}
	// }()

	// err = connection.WriteMessage(websocket.TextMessage, []byte(rawMessage))
	// if err != nil {
	// 	log.Println("write:", err)
	// 	return
	// }

	// scanner := bufio.NewScanner(os.Stdin)
	// fmt.Println("What message would you like to send??")
	// for scanner.Scan() {
	// 	fmt.Println("Writting ", scanner.Text())
	// 	connection.Write(append(scanner.Bytes(), '\r'))

	// 	fmt.Println("what message would you like to send??")
	// 	buffer := make([]byte, 1024)

	// 	_, err := connection.Read(buffer)
	// 	if err != nil && err != io.EOF {
	// 		log.Fatal(err)
	// 	} else if err == io.EOF {
	// 		log.Println("Connection is closed!")
	// 		return nil
	// 	}
	// 	fmt.Println(string(buffer))
	// }
	// return scanner.Err()

	client := &Client{socket: connection}

	cdx, err := client.socket.Write(rawMessage)
	if err != nil {
		fmt.Println("Error in sending request: ", err)
	}

	fmt.Println("write msg:", cdx)
	// defer client.socket.Close()
	// go client.send()
	// go client.receive()
	// for {
	// 	reader := bufio.NewReader(os.Stdin)
	// 	message, _ := reader.ReadString('\n')
	// 	connection.Write([]byte(strings.TrimRight(message, "\n")))
	// }
}

func NewRouter() *mux.Router {
	fmt.Println("WHY NOT start")
	r := mux.NewRouter()
	r.HandleFunc("/IsoData", IsoData).Methods("POST", "OPTIONS")
	return r

}

func IsoData(w http.ResponseWriter, r *http.Request) {
	fmt.Println("IsoData fuction started")

	var data Data

	w.Header().Set("Content-Type", "application/json")
	// json.NewDecoder(r.Body).Decode(&data)

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		fmt.Println("Error in parsing", err)
		return
	}

	fmt.Println("Welcome to ISO-8 ", data.SA)
	fmt.Println("Welcome to ", data.PAN)
	fmt.Println("Welcome ", data.BA)

	fmt.Println("Welcome to ISO-8583 data format ")
	spec := &iso8583.MessageSpec{
		Fields: map[int]field.Field{

			0: field.NewString(&field.Spec{
				Length:      4,
				Description: "Message Type Indicator",
				Enc:         encoding.ASCII,
				Pref:        prefix.ASCII.Fixed,
			}),
			1: field.NewBitmap(&field.Spec{
				Length:      16,
				Description: "Bitmap",
				Enc:         encoding.BytesToASCIIHex,
				Pref:        prefix.Hex.Fixed,
			}),
			2: field.NewString(&field.Spec{
				Length:      19,
				Description: "Primary Account Number",
				Enc:         encoding.ASCII,
				Pref:        prefix.ASCII.LL,
			}),
			3: field.NewNumeric(&field.Spec{
				Length:      6,
				Description: "Processing Code",
				Enc:         encoding.ASCII,
				Pref:        prefix.ASCII.Fixed,
				Pad:         padding.Left('0'),
			}),
			4: field.NewString(&field.Spec{
				Length:      12,
				Description: "Transaction Amount",
				Enc:         encoding.ASCII,
				Pref:        prefix.ASCII.Fixed,
				Pad:         padding.Left('0'),
			}),
			5: field.NewString(&field.Spec{
				Length:      12,
				Description: "Settlement Amount",
				Enc:         encoding.ASCII,
				Pref:        prefix.ASCII.Fixed,
				Pad:         padding.Left('0'),
			}),
			6: field.NewString(&field.Spec{
				Length:      12,
				Description: "Billing Amount",
				Enc:         encoding.ASCII,
				Pref:        prefix.ASCII.Fixed,
				Pad:         padding.Left('0'),
			}),
			// 43: field.NewComposite(&field.Spec{
			// 	Length:      150,
			// 	Description: "Card Acceptor Name/Location",
			// 	// Enc:         encoding.ASCII,
			// 	Pref: prefix.ASCII.Fixed,
			// 	Tag: &field.TagSpec{
			// 		Sort: sort.StringsByInt,
			// 	},
			// 	Subfields: map[string]field.Field{
			// 		"1": field.NewString(&field.Spec{
			// 			Length:      4,
			// 			Description: "Name of Card Acceptor",
			// 			Enc:         encoding.ASCII,
			// 			Pref:        prefix.ASCII.Fixed,
			// 		}),
			// 		"2": field.NewString(&field.Spec{
			// 			Length:      4,
			// 			Description: "City",
			// 			Enc:         encoding.ASCII,
			// 			Pref:        prefix.ASCII.Fixed,
			// 		}),
			// 		"3": field.NewString(&field.Spec{
			// 			Length:      4,
			// 			Description: "State",
			// 			Enc:         encoding.ASCII,
			// 			Pref:        prefix.ASCII.Fixed,
			// 		}),
			// 		"4": field.NewString(&field.Spec{
			// 			Length:      4,
			// 			Description: "Country Code",
			// 			Enc:         encoding.ASCII,
			// 			Pref:        prefix.ASCII.Fixed,
			// 		}),
			// 	},
			// }),
			// 7: field.NewString(&field.Spec{
			// 	Length:      10,
			// 	Description: "Transmission Date & Time",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 8: field.NewString(&field.Spec{
			// 	Length:      8,
			// 	Description: "Billing Fee Amount",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 9: field.NewString(&field.Spec{
			// 	Length:      8,
			// 	Description: "Settlement Conversion Rate",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 10: field.NewString(&field.Spec{
			// 	Length:      8,
			// 	Description: "Cardholder Billing Conversion Rate",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 11: field.NewString(&field.Spec{
			// 	Length:      6,
			// 	Description: "Systems Trace Audit Number (STAN)",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 12: field.NewString(&field.Spec{
			// 	Length:      6,
			// 	Description: "Local Transaction Time",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 13: field.NewString(&field.Spec{
			// 	Length:      4,
			// 	Description: "Local Transaction Date",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 14: field.NewString(&field.Spec{
			// 	Length:      4,
			// 	Description: "Expiration Date",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 15: field.NewString(&field.Spec{
			// 	Length:      4,
			// 	Description: "Settlement Date",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 16: field.NewString(&field.Spec{
			// 	Length:      4,
			// 	Description: "Currency Conversion Date",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 17: field.NewString(&field.Spec{
			// 	Length:      4,
			// 	Description: "Capture Date",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 18: field.NewString(&field.Spec{
			// 	Length:      4,
			// 	Description: "Merchant Type",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 19: field.NewString(&field.Spec{
			// 	Length:      3,
			// 	Description: "Acquiring Institution Country Code",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 20: field.NewString(&field.Spec{
			// 	Length:      3,
			// 	Description: "PAN Extended Country Code",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 21: field.NewString(&field.Spec{
			// 	Length:      3,
			// 	Description: "Forwarding Institution Country Code",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 22: field.NewString(&field.Spec{
			// 	Length:      3,
			// 	Description: "Point of Sale (POS) Entry Mode",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 23: field.NewString(&field.Spec{
			// 	Length:      3,
			// 	Description: "Card Sequence Number (CSN)",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 24: field.NewString(&field.Spec{
			// 	Length:      3,
			// 	Description: "Function Code",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 25: field.NewString(&field.Spec{
			// 	Length:      2,
			// 	Description: "Point of Service Condition Code",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 26: field.NewString(&field.Spec{
			// 	Length:      2,
			// 	Description: "Point of Service PIN Capture Code",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 27: field.NewString(&field.Spec{
			// 	Length:      1,
			// 	Description: "Authorizing Identification Response Length",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 28: field.NewString(&field.Spec{
			// 	Length:      9,
			// 	Description: "Transaction Fee Amount",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 29: field.NewString(&field.Spec{
			// 	Length:      9,
			// 	Description: "Settlement Fee Amount",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 30: field.NewString(&field.Spec{
			// 	Length:      9,
			// 	Description: "Transaction Processing Fee Amount",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 31: field.NewString(&field.Spec{
			// 	Length:      9,
			// 	Description: "Settlement Processing Fee Amount",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 32: field.NewString(&field.Spec{
			// 	Length:      11,
			// 	Description: "Acquiring Institution Identification Code",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.LL,
			// }),
			// 33: field.NewString(&field.Spec{
			// 	Length:      11,
			// 	Description: "Forwarding Institution Identification Code",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.LL,
			// }),
			// 34: field.NewString(&field.Spec{
			// 	Length:      28,
			// 	Description: "Extended Primary Account Number",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.LL,
			// }),
			// 35: field.NewString(&field.Spec{
			// 	Length:      37,
			// 	Description: "Track 2 Data",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.LL,
			// }),
			// 36: field.NewString(&field.Spec{
			// 	Length:      104,
			// 	Description: "Track 3 Data",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.LLL,
			// }),
			// 37: field.NewString(&field.Spec{
			// 	Length:      12,
			// 	Description: "Retrieval Reference Number",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 38: field.NewString(&field.Spec{
			// 	Length:      6,
			// 	Description: "Authorization Identification Response",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 39: field.NewString(&field.Spec{
			// 	Length:      2,
			// 	Description: "Response Code",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 40: field.NewString(&field.Spec{
			// 	Length:      3,
			// 	Description: "Service Restriction Code",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 41: field.NewString(&field.Spec{
			// 	Length:      8,
			// 	Description: "Card Acceptor Terminal Identification",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 42: field.NewString(&field.Spec{
			// 	Length:      15,
			// 	Description: "Card Acceptor Identification Code",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 43: field.NewComposite(&field.Spec{
			// 	Length:      150,
			// 	Description: "Card Acceptor Name/Location",
			// 	// Enc:         encoding.ASCII,
			// 	Pref: prefix.ASCII.Fixed,
			// 	Tag: &field.TagSpec{
			// 		Sort: sort.StringsByInt,
			// 	},
			// 	Subfields: map[string]field.Field{
			// 		"1": field.NewString(&field.Spec{
			// 			Length:      4,
			// 			Description: "Name of Card Acceptor",
			// 			Enc:         encoding.ASCII,
			// 			Pref:        prefix.ASCII.Fixed,
			// 		}),
			// 		"2": field.NewString(&field.Spec{
			// 			Length:      4,
			// 			Description: "City",
			// 			Enc:         encoding.ASCII,
			// 			Pref:        prefix.ASCII.Fixed,
			// 		}),
			// 		"3": field.NewString(&field.Spec{
			// 			Length:      4,
			// 			Description: "State",
			// 			Enc:         encoding.ASCII,
			// 			Pref:        prefix.ASCII.Fixed,
			// 		}),
			// 		"4": field.NewString(&field.Spec{
			// 			Length:      4,
			// 			Description: "Country Code",
			// 			Enc:         encoding.ASCII,
			// 			Pref:        prefix.ASCII.Fixed,
			// 		}),
			// 	},
			// }),
			// 44: field.NewString(&field.Spec{
			// 	Length:      99,
			// 	Description: "Additional Data",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.LL,
			// }),
			// 45: field.NewString(&field.Spec{
			// 	Length:      76,
			// 	Description: "Track 1 Data",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.LL,
			// }),
			// 46: field.NewString(&field.Spec{
			// 	Length:      999,
			// 	Description: "Additional data (ISO)",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.LLL,
			// }),
			// 47: field.NewString(&field.Spec{
			// 	Length:      999,
			// 	Description: "Additional data (National)",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.LLL,
			// }),
			// 48: field.NewString(&field.Spec{
			// 	Length:      999,
			// 	Description: "Additional data (Private)",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.LLL,
			// }),
			// 49: field.NewString(&field.Spec{
			// 	Length:      3,
			// 	Description: "Transaction Currency Code",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 50: field.NewString(&field.Spec{
			// 	Length:      3,
			// 	Description: "Settlement Currency Code",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 51: field.NewString(&field.Spec{
			// 	Length:      3,
			// 	Description: "Cardholder Billing Currency Code",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 52: field.NewString(&field.Spec{
			// 	Length:      8,
			// 	Description: "PIN Data",
			// 	Enc:         encoding.BytesToASCIIHex,
			// 	Pref:        prefix.Hex.Fixed,
			// }),
			// 53: field.NewString(&field.Spec{
			// 	Length:      16,
			// 	Description: "Security Related Control Information",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
			// 54: field.NewString(&field.Spec{
			// 	Length:      120,
			// 	Description: "Additional Amounts",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.LLL,
			// }),
			// 55: field.NewString(&field.Spec{
			// 	Length:      999,
			// 	Description: "ICC Data – EMV Having Multiple Tags",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.LLL,
			// }),
			// 56: field.NewString(&field.Spec{
			// 	Length:      999,
			// 	Description: "Reserved (ISO)",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.LLL,
			// }),
			// 57: field.NewString(&field.Spec{
			// 	Length:      999,
			// 	Description: "Reserved (National)",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.LLL,
			// }),
			// 58: field.NewString(&field.Spec{
			// 	Length:      999,
			// 	Description: "Reserved (National)",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.LLL,
			// }),
			// 59: field.NewString(&field.Spec{
			// 	Length:      999,
			// 	Description: "Reserved (National)",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.LLL,
			// }),
			// 60: field.NewString(&field.Spec{
			// 	Length:      999,
			// 	Description: "Reserved (National)",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.LLL,
			// }),
			// 61: field.NewString(&field.Spec{
			// 	Length:      999,
			// 	Description: "Reserved (Private)",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.LLL,
			// }),
			// 62: field.NewString(&field.Spec{
			// 	Length:      999,
			// 	Description: "Reserved (Private)",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.LLL,
			// }),
			// 63: field.NewString(&field.Spec{
			// 	Length:      999,
			// 	Description: "Reserved (Private)",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.LLL,
			// }),
			// 64: field.NewString(&field.Spec{
			// 	Length:      8,
			// 	Description: "Message Authentication Code (MAC)",
			// 	Enc:         encoding.BytesToASCIIHex,
			// 	Pref:        prefix.Hex.Fixed,
			// }),
			// 90: field.NewString(&field.Spec{
			// 	Length:      42,
			// 	Description: "Original Data Elements",
			// 	Enc:         encoding.ASCII,
			// 	Pref:        prefix.ASCII.Fixed,
			// }),
		},
	}
	// create message with defined spec
	message := iso8583.NewMessage(spec)

	// sets message type indicator at field 0
	message.MTI("0100")

	// set all message fields you need as strings
	//message.Field(1, "673333333333")
	message.Field(2, data.PAN) //PAN number
	message.Field(3, fmt.Sprint(data.PC))
	message.Field(4, data.TA)
	message.Field(5, data.SA)

	message.Field(6, data.BA)

	// F4:  field.NewNumericValue(77700),
	// 		F7:  field.NewNumericValue(701111844),
	// 		F11: field.NewNumericValue(123),
	// 		F12: field.NewNumericValue(131844),
	// 		F13: field.NewNumericValue(701),
	// message.Field(2,fmt.Sprint(field.Spec().Subfields(map[1])))

	// message.Bitmap().Spec().Subfields(1,"name",)

	// generate binary representation of the message into rawMessage
	rawMessage, err := message.Pack()
	if err != nil {
		fmt.Println("error is ", err)
	}
	fmt.Println("Raw message is ", string(rawMessage))

	flagMode := flag.String("mode", data.Mode, "start in client or server mode")
	flag.Parse()
	if strings.ToLower(*flagMode) == "client" {
		startClientMode(rawMessage)
	} else {
		startServerMode()
	}

	w.Write([]byte("Raw message is " + string(rawMessage)))

	// jsonMessage, err := json.Marshal(message)
	// if err != nil {
	// 	fmt.Println("error is ", err)
	// }
	// fmt.Println("Raw message is for json marshal ", string(jsonMessage), jsonMessage)

	// type ISO87Data struct {
	// 	F2 *field.String
	// 	F3 *field.Numeric
	// 	F4 *field.String
	// 	F5 *field.String
	// 	F6 *field.String
	// 	// F43 *field.Composite
	// }

	// message.SetData(&ISO87Data{})

	// // let's unpack binary message

	// err = message.Unpack(rawMessage)
	// if err != nil {
	// 	fmt.Println("error is ", err)
	// }

	// // err = message.UnmarshalJSON(jsonMessage)
	// // if err != nil {
	// // 	fmt.Println("error is ", err)
	// // }

	// // to get access to typed data we have to get Data from the message and convert it into our ISO87Data type
	// data1 := message.Data().(*ISO87Data)

	// // now you have typed values
	// fmt.Println(data1.F2.Value) // is a string "4242424242424242"
	// fmt.Println(data1.F3.Value) // is an int 123456
	// fmt.Println(data1.F4.Value) // is a string "100"
	// fmt.Println(data1.F5.Value)
	// fmt.Println(data1.F6.Value)
	// fmt.Println(data1.F6.Spec().Description)
	// fmt.Println("con start")

	// fmt.Println(data1.F43.Spec().Tag)

}
