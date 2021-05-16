package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

//Мультиплексирование потоков данных STS-1 играет важную роль в технологии SONET.
//Мультиплексор 3:1 уплотняет три входных потока STS-1 в один выходной поток STS-3.
//Уплотнение производится побайтно, то есть первые три выходных байта соответствуют первым байтам входных потоков 1, 2 и 3 соответственно.
//Следующие три байта — вторым байтам потоков 1, 2 и 3 и т. д.
//Напишите программу, симулирующую работу мультиплексора 3:1. В программе должно быть пять процессов.
//Главный создает четыре других процесса (для трех входных потоков и мультиплексора).
//Каждый процесс входного потока считывает в кадр STS-1 данные из файла в виде последовательности из 810 байт.
//Затем кадры побайтно отсылаются процессу мультиплексора.
//Мультиплексор принимает потоки и выводит результирующий кадр STS-3 (снова побайтно), записывая его на стандартное устройство вывода.
//Для взаимодействия между процессами используйте метод конвейеров (pipes).

const (
	frameSize    = 810
	numOfThreads = 3
)

func reader(fileName string, out chan<- byte) {
	defer func() {
		close(out)
	}()

	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		panic(err)
	}

	// Считываем данные по фреймам в 810 байт
	for i := int64(0); i < info.Size(); i += frameSize {
		// Готовим фрейм
		frame := make([]byte, frameSize)

		// Читаем файл и смотрим на ошибку, если файл вычитан (io.EOF), горутина прекращает свою работу
		if _, err := file.Read(frame); err != nil {
			switch {
			case errors.Is(err, io.EOF):
				return
			default:
				panic(err)
			}
		}

		// Побайтовая передача данных в нужный канал
		for j := 0; j < len(frame); j++ {
			out <- frame[j]
		}
	}
}

func multiplexer(outChannels []chan byte, multiplexedDataCh chan<- []byte) {
	data := make([]byte, 0)

	// Пока хотя бы один канал отдает полезную информацию, мы читаем из всех
	// Закрытые каналы отдадут нам zero value, что при разложении мультиплексированных данных легко фильтруется
	for {
		atLeastOneChIsOpen := false

		for i := 0; i < numOfThreads; i++ {
			b, ok := <-outChannels[i]
			if ok {
				atLeastOneChIsOpen = true
			}

			data = append(data, b)
		}

		if !atLeastOneChIsOpen {
			break
		}
	}

	fmt.Println(data)
	multiplexedDataCh <- data
}

func main() {
	outChannels := make([]chan byte, 0, numOfThreads)
	for i := 0; i < numOfThreads; i++ {
		outChannels = append(outChannels, make(chan byte, frameSize))
	}

	for i := 0; i < numOfThreads; i++ {
		fileName := fmt.Sprintf("./physical-layer/multiplexer/data/file%d.in", i+1)
		go reader(fileName, outChannels[i])
	}

	multiplexedDataCh := make(chan []byte)
	go multiplexer(outChannels, multiplexedDataCh)

	//Ожидаем завершения работы мультиплексора и восстанавливаем данные для проверки правильности работы мультиплексера
	recoverData(<-multiplexedDataCh)
}

func recoverData(data []byte) {
	fmt.Println("Returning data back...")

	recoveredData := make([][]byte, 0)
	for i := 0; i < numOfThreads; i++ {
		recoveredData = append(recoveredData, make([]byte, 0))
	}

	for i := 0; i < len(data); i += numOfThreads {
		// Читаем по 3 байта сразу и пишем их в нужные массивы байт для последующего восстановления данных
		for j := 0; j < numOfThreads; j++ {
			// Один из 3 байт может быть пустым (см. функцию reader), и его не надо писать
			if data[i+j] != 0 {
				recoveredData[j] = append(recoveredData[j], data[i+j])
			}
		}
	}

	// Сохраняем файлы и их данные
	for i, d := range recoveredData {
		saveRecoveredData(i+1, d)
	}

	fmt.Println("Data successfully recovered!")
}

func saveRecoveredData(num int, data []byte) {
	file, err := os.Create(fmt.Sprintf("./physical-layer/multiplexer/data/file%d.out", num))
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if _, err := file.WriteString(string(data)); err != nil {
		panic(err)
	}
}
