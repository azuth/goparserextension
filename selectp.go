package main

import "time"
import "fmt"

func main() {

	ca0 := make(chan string)
	ca1 := make(chan string)
	cb0 := make(chan string)
	cb1 := make(chan string)
	cc0 := make(chan string)
	cc1 := make(chan string)

	go func() {
		ca0 <- "a0"
	}()

	go func() {
		ca1 <- "a1"
	}()

	go func() {
		cb0 <- "b0"
	}()

	go func() {
		cb1 <- "b1"
	}()

	go func() {
		cc0 <- "c0"
	}()

	go func() {
		cc1 <- "c1"
	}()

	time.Sleep(time.Second * 1)

	select {
	case msg1 := <-ca0:
		fmt.Println("received", msg1)
	case msg2 := <-ca1:
		fmt.Println("received", msg2)
	}

Loop1:
	for {
		select {
		case msg1 := <-cb0:
			fmt.Println("received", msg1)
			break Loop1
		default:
			select {
			case msg2 := <-cb1:
				fmt.Println("received", msg2)
				break Loop1
			default:
			}
		}
	}

Loop2:
	selectp {
	case msg1 := <-cc0:
		fmt.Println("received", msg1)
		break Loop2
	case msg2 := <-cc1:
		fmt.Println("received", msg2)
		break Loop2
	}
}
