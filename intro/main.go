package main

import (
	"bufio"
	"fmt"
	"os"
)

// Напишите программу, которая осуществляет поток сообщений от верхнего слоя до нижнего слоя модели протокола с 7 уровнями.
// Ваша программа должна включать отдельную функцию протокола для каждого уровня.
// Заголовки протокола — последовательность до 64 символов.
// У каждой функции есть два параметра: сообщение, пришедшее из протокола более высокого уровня (буфер случайной работы),
// и размер сообщения. Эта функция присоединяет свой заголовок перед сообщением, печатает новое сообщение на
// стандартном выводе и затем вызывает функцию протокола нижнего уровня. Ввод программы — сообщение приложения
// (последовательность 80 символов или меньше).

// OSI модель
type OSI struct {
	layers [7]string
	curLvl uint8
}

// NewOSI конструктор модели OSI
func NewOSI() *OSI {
	return &OSI{
		layers: [7]string{"Physical", "Data", "Network", "Transport", "Session", "Presentation", "Application"},
		curLvl: 0,
	}
}

// AcceptMessage принимает любое небольшое сообщение и добавляет к нему заголовки всех слоев модели OSI
func (o *OSI) AcceptMessage(msg string) {
	if len(msg) > 80 {
		panic("Can't accept message with length more than 80 characters")
	}

	fmt.Println(o.NextLvl(msg))
}

// NextLvl - рекурсивное погружение по слоям модели OSI, чтобы не хардкодить 1 слой - 1 метод
func (o *OSI) NextLvl(msg string) string {
	nextMsg := fmt.Sprintf("%s: %s", o.layers[o.curLvl], msg)

	if o.curLvl < 6 {
		o.curLvl++
		return o.NextLvl(nextMsg)
	}

	o.curLvl = 0
	return nextMsg
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter message: ")

	text, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}

	osi := NewOSI()
	osi.AcceptMessage(text)
}
