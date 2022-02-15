package main

func Fanout(In <-chan string, Out1, Out2 chan string) {
	//take an input channel and fan it out to multiple output channels
	for v := range In { // recieve until closed
		select { // send to whichever channel is ready (non-blocking)
		case Out1 <- v:
		case Out2 <- v:
		}
	}
	close(Out1)
	close(Out2)
}
