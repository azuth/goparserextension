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

	{
	}

	selectp {
	case msg1 := <-cb0:
		fmt.Println("received", msg1)
	case msg2 := <-cb1:
		fmt.Println("received", msg2)
	case msg3 := <-cb2:
		fmt.Println("received", msg3)
	}

	selectp {
	case msg1 := <-cb0:
		fmt.Println("received", msg1)
		fmt.Println("received", msg1)
		fmt.Println("received", msg1)
	case msg2 := <-cb1:
		fmt.Println("received", msg2)
		fmt.Println("received", msg2)
		fmt.Println("received", msg2)
	case msg3 := <-cb2:
		fmt.Println("received", msg3)
		fmt.Println("received", msg3)
		fmt.Println("received", msg3)
	default:
		fmt.Println("and again")
		fmt.Println("and again")
		fmt.Println("and again")
	}

	selectp {
	case msg1 := <-cb0:
		fmt.Println("received", msg1)
	case msg2 := <-cb1:
		selectp {
		case msg2 := <-cb1:
			fmt.Println("received", msg2)

		default:
			fmt.Println("and again")
		}

	default:
		selectp {
		case msg2 := <-cb1:
			fmt.Println("received", msg2)

		default:
			fmt.Println("and again")
		}
	}

Loop2:
	selectp {
	case msg1 := <-cc0:
		fmt.Println("received", msg1)
		break Loop2
	case msg2 := <-cc1:
		fmt.Println("received", msg2)
	default:
		selectp {
		case msg1 := <-cc0:
			fmt.Println("received", msg1)
			//		goto Loop2
		case msg2 := <-cc1:
			fmt.Println("received", msg2)
		}
	}

	selectp {
	case msg1 := <-cc0:
		fmt.Println("received", msg1)
	default:
		fmt.Print("I am so empty")
	}
}
