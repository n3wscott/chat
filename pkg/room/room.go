package room

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/n3wscott/chat/pkg/api"
	"github.com/n3wscott/chat/pkg/client"
	"log"
	"strings"
	"time"
)

func NewRoom(me string, host string, port int) *Room {
	r := &Room{
		me:    me,
		here:  []string(nil),
		Done:  make(chan bool, 1),
		Entry: make(chan api.Message, 10),
		Room:  make(chan api.Message, 10),
	}
	go r.connectClient(host, port)
	return r
}

func (r *Room) connectClient(host string, port int) {
	c := client.NewClient(r.me, host, port)

	go c.Run()
	defer func() {
		c.Done <- true
	}()

	for {
		select {
		case msg := <-r.Entry:
			m := api.Message{Author: r.me, Body: msg.Body}
			c.Tx <- m
		case m := <-c.Msg:
			r.Room <- m
		case h := <-c.Here:
			if r.addHere(h) {
				r.Room <- api.Message{Body: fmt.Sprintf("%s has joined the room.", h)}
			}
		}
	}
}

func (r *Room) addHere(n string) bool {
	for _, h := range r.here {
		if h == n {
			return false
		}
	}
	r.here = append(r.here, n)
	return true
}

const (
	ColorYellow = 226
	ColorGreen  = 2
)

type Room struct {
	me    string
	here  []string
	Done  chan bool
	Entry chan api.Message
	Room  chan api.Message
}

func setCurrentViewOnTop(g *gocui.Gui, name string) (*gocui.View, error) {
	if _, err := g.SetCurrentView(name); err != nil {
		return nil, err
	}
	return g.SetViewOnTop(name)
}

func (r *Room) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("here", -1, -1, 20, maxY-2); err != nil &&
		err != gocui.ErrUnknownView {
		return err
	} else {
		v.Clear()
		for _, h := range r.here {
			fmt.Fprintln(v, h)
		}
	}
	if v, err := g.SetView("chat", 20, -1, maxX, maxY-2); err != nil &&
		err != gocui.ErrUnknownView {
		return err
	} else {
		v.Wrap = true
		v.Autoscroll = true
	}
	if v, err := g.SetView("input", -1, maxY-2, maxX, maxY); err != nil &&
		err != gocui.ErrUnknownView {
		return err
	} else {
		v.Title = r.me
		v.Editable = true
		v.Wrap = true
		v.Autoscroll = true

		if _, err := setCurrentViewOnTop(g, "input"); err != nil {
			return err
		}
		g.Cursor = true
		v.Highlight = true
		g.SetViewOnTop("input")
	}
	/*
		if v, err := g.SetView("but1", 2, 2, 22, 7); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Highlight = true
			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack
			fmt.Fprintln(v, "Button 1 - line 1")
			fmt.Fprintln(v, "Button 1 - line 2")
			fmt.Fprintln(v, "Button 1 - line 3")
			fmt.Fprintln(v, "Button 1 - line 4")
		}
		if v, err := g.SetView("but2", 24, 2, 44, 4); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Highlight = true
			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack
			fmt.Fprintln(v, "Button 2 - line 1")
		}
	*/
	return nil
}

func (r *Room) quit(g *gocui.Gui, v *gocui.View) error {
	log.Print("got to done")
	r.Done <- true
	return gocui.ErrQuit
}

func (r *Room) post(g *gocui.Gui, v *gocui.View) error {

	r.Entry <- api.Message{
		Body: v.Buffer(),
	}

	v.Clear()
	v.SetCursor(0, 0)

	return nil
}

func showMsg(g *gocui.Gui, v *gocui.View) error {
	var l string
	var err error

	if _, err := g.SetCurrentView(v.Name()); err != nil {
		return err
	}

	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}

	maxX, maxY := g.Size()
	if v, err := g.SetView("msg", maxX/2-10, maxY/2, maxX/2+10, maxY/2+2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, l)
	}
	return nil
}

func delMsg(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView("msg"); err != nil {
		return err
	}
	return nil
}

func (r *Room) keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, r.quit); err != nil {
		return err
	}
	for _, n := range []string{"but1", "but2"} {
		if err := g.SetKeybinding(n, gocui.MouseLeft, gocui.ModNone, showMsg); err != nil {
			return err
		}
	}
	if err := g.SetKeybinding("msg", gocui.MouseLeft, gocui.ModNone, delMsg); err != nil {
		return err
	}

	if err := g.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, r.post); err != nil {
		return err
	}

	if v, err := g.View("side"); err == nil {
		fmt.Fprintln(v, "bot")
		fmt.Fprintln(v, "John")
		fmt.Fprintln(v, "Jim")
		fmt.Fprintln(v, "@me")
	}

	return nil
}

func (r *Room) Run() *Room {
	ready := make(chan bool, 1)

	g, err := gocui.NewGui(gocui.Output256)
	if err != nil {
		log.Panicln(err)
	}

	go func() {
		time.Sleep(400 * time.Millisecond)
		ready <- true
	}()

	go func() {
		g.Cursor = true
		g.Mouse = true
		g.SelFgColor = gocui.ColorGreen

		g.SetManagerFunc(r.layout)

		if err := r.keybindings(g); err != nil {
			log.Panicln(err)
		}

		if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
			r.Done <- true
		}
		g.Close()
	}()

	go func() {
		for {
			select {
			case msg := <-r.Room:
				chat, err := g.View("chat")
				if err != nil {
					continue
				}
				if strings.HasSuffix(msg.Body, "\n") {
					msg.Body = strings.TrimSuffix(msg.Body, "\n")
				}
				if len(msg.Author) == 0 {
					fmt.Fprintf(chat, "\x1b[38;5;%dm%s\x1b[0m\n", ColorYellow, msg.Body)

					g.Update(func(gui *gocui.Gui) error { return nil })

				} else if len(msg.Body) > 0 {
					if msg.Author == r.me {
						fmt.Fprintf(chat, "\x1b[38;5;%dm%s\x1b[0m: %s\n", ColorGreen, msg.Author, msg.Body)
					} else {
						fmt.Fprintf(chat, "%s: %s\n", msg.Author, msg.Body)
					}
					g.Update(func(gui *gocui.Gui) error { return nil })
				}
			}
		}
	}()

	<-ready
	return r
}
