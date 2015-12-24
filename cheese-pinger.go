package main


import (
	"database/sql"
	"net"
	"os"
	"log"
	"fmt"
	"time"
	"errors"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/internal/iana"
	_ "github.com/lib/pq"
)

func resolver(host string) string {
	var htpIP []string
	htpIP, err := net.LookupHost(host)
	
	if err != nil {
		return "error"
	} else {
		return htpIP[0]
	}
	
}

func pingMyHost(htp string) (string,string, error)  {
	
	if (resolver(htp) == "error"){
		return "Cant resolve host",htp,errors.New("Resolve error!")
	} else {
		log.Printf("%v resolved. Sending cheese!",htp)
		targetIP := resolver(htp)
		c,err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
		if err != nil {
			return "Socket listen error",htp,err
		}	
		defer c.Close()
		
		wm := icmp.Message{
			Type: ipv4.ICMPTypeEcho, Code: 0,
			Body: &icmp.Echo{
				ID: os.Getpid() & 0xffff, Seq: 1,
				Data: []byte("Sending GORGONZOLLA"),
			},	
		}
		
		wb, err := wm.Marshal(nil)
		if err != nil {
			return "Marshal error",htp, err
		}
		
		if _, err := c.WriteTo(wb, &net.IPAddr{IP: net.ParseIP(targetIP)}); err != nil {
			return "WriteTo error",htp, err
		}
		
		rb := make([]byte, 1500)
		
		// set timeout to 3 secs
		if err := c.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
			return "SetReadDeadline error",htp,err
		}
		n, _, err := c.ReadFrom(rb)
		if err != nil {
			return "No response. Cheese is lost!",htp, err
		}
		
		rm, err := icmp.ParseMessage(iana.ProtocolICMP, rb[:n])
		if err != nil {
			return "ParseMessage error",htp, err
		}
		switch rm.Type{
		case ipv4.ICMPTypeEchoReply:
			return "DELIVERY - Success",htp,err
		default:
			return "Cheese is lost!",htp,errors.New("No answer from remote host")
		}
	}
}

func main(){

	// db connect
	db,err := sql.Open("postgres", "host= port= dbname= user= password= sslmode=verify-full")
	if err != nil {
		log.Fatal("! DB connection error: ",err)
	}
	
	//get rows from db
	rows, err := db.Query("SELECT * FROM table") // put your query here
	if err != nil {
		log.Fatal("!Error! Cant aquire rows from DB ",err)
	}
	
	//parse rows
	cols, err := rows.Columns()
	colLen := len(cols)
	if err != nil {
		log.Fatal("!Error! No columns detected.", err);
	}
	
	var fields []interface{}
	for i:=0;i<colLen;i++{
		fields = append(fields,new(string))
	}
	
	for rows.Next(){
		rows.Scan(fields...)
		response,host,err :=pingMyHost(*(fields[0].(*string)))
		if err != nil {
			fmt.Println("=RESP=========")
			log.Printf("!Awww! %v", response)
			log.Printf("-> %v", host)
			log.Printf("Description: %v",err)
			fmt.Println()
		} else {
			fmt.Println("=RESP=========")
			log.Printf("WOOOO! %v", response)
			log.Printf("-> %v",host)
			fmt.Println()
		}
	}
}
