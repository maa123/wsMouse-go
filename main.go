package main

import(
	"syscall"
	"net/http"
	"image"
	"strings"
	"strconv"
	"net"
	"golang.org/x/net/websocket"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/skip2/go-qrcode"
)

var(
	user32, _ =   syscall.LoadDLL("user32.dll")
	mouse_event, _ = user32.FindProc("mouse_event")
)

func moveMouse(x int, y int){
	mouse_event.Call(uintptr(0x01), uintptr(x), uintptr(y), 0, 0)
}

func strToFloat(str string) float64 {
	ret, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return float64(0)
	}
	return ret
}

func receiveMsg(msg string) {
	if(msg == "cd"){
		mouse_event.Call(uintptr(0x02), 0, 0, 0, 0)
	}else if(msg == "cu"){
		mouse_event.Call(uintptr(0x04), 0, 0, 0, 0)
	}else{
		var msgs []string = strings.Split(msg, ",")
		if(msgs[0] == "p"){
			var x float64 = strToFloat(msgs[1])
			var y float64 = strToFloat(msgs[2])
			moveMouse(int(-x/2.0), int(-y/2.0))
		}else if(msgs[0] == "m"){
			mouse_event.Call(uintptr(0x800), 0, 0, uintptr(-int(strToFloat(msgs[2]))), 0)
		}
	}
}

func wsServe(ws *websocket.Conn){
	var err error
	defer ws.Close()
	var msg string
	for {
		if err = websocket.Message.Receive(ws, &msg); err != nil {
			break
		}
		receiveMsg(msg)
	}
}

func wsFunc(w http.ResponseWriter, r *http.Request) {
	s := websocket.Server{Handler: websocket.Handler(wsServe)}
	s.ServeHTTP(w, r)
}

func getMyLocalIP() string {
	addrs, _ := net.InterfaceAddrs()
	for _, a := range addrs {
		if ipn, ct := a.(*net.IPNet); ct && !ipn.IP.IsLoopback() {
			if (ipn.IP.To4() != nil) && !ipn.IP.IsLinkLocalUnicast() {
				return ipn.IP.String()
			}
		}
	}
	return "127.0.0.1"
}
func startServe() {
	http.HandleFunc("/", wsFunc)
	err := http.ListenAndServe(":8765", nil)
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	defer user32.Release()
	imgv := new(walk.ImageView)
	q, err := qrcode.New("ws://"+getMyLocalIP()+":8765", qrcode.Medium)
	if err != nil {
		panic(err.Error())
	}
	var qri image.Image = q.Image(160)
	qim, er := walk.NewBitmapFromImage(qri)
	if er != nil {
		panic(er.Error())
	}
	go startServe()
	MainWindow{
		Title:   "TrackPad Bridge",
		MinSize: Size{240, 200},
		Size: Size{320, 200},
		Layout:  VBox{MarginsZero: true},
		Children: []Widget{
			ImageView{
				AssignTo: &imgv,
				MinSize: Size{180, 180},
				MaxSize: Size{180, 180},
				Image: qim,
			},
		},
	}.Run()
}