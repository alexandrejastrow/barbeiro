package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
)

func barb(id int, val int, leitor chan int, escritor chan int, cadCliente []chan int,
	cabeloCortado chan int, myCad chan int, stop chan int, msg chan string) {

	count := 0
	var flag bool = true
	for {
		<-escritor
		leitor <- id
		<-cadCliente[id]
		cabeloCortado <- 1
		count++

		if count >= val {
			if flag {
				str := fmt.Sprintf("barbeiro %d  cortou %d vezes", id, count)
				stop <- 1
				msg <- str
				flag = false
			}
		}

	}
}

func cli(leitor chan int, escritor chan int, cad chan int, cadBarbeiro []chan int,
	cadCliente []chan int, cabCortado []chan int, newClient chan int) {

	select {
	case <-cad:

		i := <-leitor
		escritor <- 1
		<-cadBarbeiro[i]
		cadCliente[i] <- i
		cad <- 1
		<-cabCortado[i]
		cadBarbeiro[i] <- 1
		newClient <- 1
		return

	default:
		newClient <- 1
	}

}

func verifiParams(s []string) (int, int, int, error) {

	a, _ := strconv.ParseInt(s[1], 10, 64)
	b, _ := strconv.ParseInt(s[2], 10, 64)
	c, _ := strconv.ParseInt(s[3], 10, 64)

	if a <= 0 || b <= 0 || c <= 0 {
		return 0, 0, 0, errors.New("error dos parametros")
	}

	return int(a), int(b), int(c), nil
}

func main() {

	args := os.Args

	if len(args) < 4 {
		os.Exit(1)
	}

	var i int
	var count int = 0

	nBarbeiros, nCadeiras, nClientes, err := verifiParams(os.Args)

	if err != nil {
		panic(err)
	}

	numGoRout := nCadeiras + nClientes
	var wg sync.WaitGroup

	var chCadeiras = make(chan int, int(nCadeiras))
	var barbeiroCadeira = []chan int{}
	var cabeloCortado = []chan int{}
	var clinteCadeira = []chan int{}
	var leitor = make(chan int)
	var escritor = make(chan int, 1)
	var newClient = make(chan int, numGoRout)
	var chStop = make(chan int, nBarbeiros)
	var msg = make(chan string, nBarbeiros)

	escritor <- 1

	wg.Add(nBarbeiros)

	for i = 0; i < nBarbeiros; i++ {

		clinteCadeira = append(clinteCadeira, make(chan int))
		barbeiroCadeira = append(barbeiroCadeira, make(chan int, 1))
		cabeloCortado = append(cabeloCortado, make(chan int))
		barbeiroCadeira[i] <- 1
		go barb(i, int(nClientes), leitor, escritor, clinteCadeira,
			cabeloCortado[i], barbeiroCadeira[i], chStop, msg)
	}
	for i = 0; i < nCadeiras; i++ {
		chCadeiras <- 1
	}

	for i = 0; i < numGoRout; i++ {
		newClient <- 1
	}

	go func(i int, msg chan string) {
		count := 0
		for i := range msg {
			fmt.Println(i)
			count++
			if count >= nBarbeiros {
				os.Exit(0)
			}
		}

	}(nBarbeiros, msg)

	for {

		<-newClient
		go cli(leitor, escritor, chCadeiras, barbeiroCadeira, clinteCadeira, cabeloCortado, newClient)
		count += 1
		if count >= numGoRout {
			count = 0
		}

	}
}
