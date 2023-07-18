package main

import (
	"bufio"
	"container/list"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type DataIn struct {
	tables      int
	costForHour int
	timeStart   time.Time
	timeEnd     time.Time
	subjects    []Subject
}

type Subject struct {
	timeOfSubj time.Time
	id         int
	name       string
	tableNumb  int
}

type DataOut struct {
	timeStart   time.Time
	timeEnd     time.Time
	payment     []Pay
	subjectsOut []Subject
}

type Pay struct {
	table       int
	money       int
	workingTime string
}

type Tables struct {
	// number int
	pay        int
	owner      string
	startOwned time.Time
	endOwned   time.Time
	timeInUsed time.Duration
}

type StackSubj[T any] struct {
	data []T
}

func (s *StackSubj[T]) Pop() T {
	if len(s.data) == 0 {
		panic("stack is empty")
	}

	last := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]

	return last
}

func (s *StackSubj[T]) Show() T {
	if len(s.data) == 0 {
		panic("stack is empty")
	}

	last := s.data[len(s.data)-1]

	return last
}

func (s *StackSubj[T]) Push(v T) {
	s.data = append(s.data, v)
}

func (s *StackSubj[T]) IsEmpty() bool {
	return len(s.data) == 0
}

func readData(path string) ([]string, error) {
	//open file txt
	text := make([]string, 0)
	file, err := os.Open(path)
	if err != nil {
		e := errors.New(fmt.Sprintf("Could not open file: %s", err))
		return nil, e
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)

	//read the data
	for fileScanner.Scan() {
		text = append(text, fileScanner.Text())
	}

	if err := fileScanner.Err(); err != nil {
		e := errors.New(fmt.Sprintf("Could not scan the file: %s", err))
		return nil, e
	}

	return text, nil
}

func formatDataIn(data []string) (DataIn, error) {
	formattingData := DataIn{}
	var err error

	//Write tables
	formattingData.tables, err = strconv.Atoi(data[0])
	if err != nil || formattingData.tables < 0 {
		e := errors.New(fmt.Sprintf("Error with tables in line '%s': %s", data[0], err))
		return DataIn{}, e
	}

	//Check for two times
	times := strings.Split(data[1], " ")
	if len(times) != 2 {
		e := errors.New(fmt.Sprintf("Error with start and end times in line '%s': %s", data[1], err))
		return DataIn{}, e
	}

	//Kostil dlya sravnenia
	var timeStart, timeEnd time.Time

	for i := 0; i < 2; i++ {
		t, err := time.Parse("15:04", times[i])
		if err != nil {
			e := errors.New(fmt.Sprintf("Error with times in line '%s': %s", data[1], err))
			return DataIn{}, e
		}
		if i == 0 {
			timeStart = t
		} else if i == 1 {
			timeEnd = t
		}
	}

	//God of errors
	if timeStart.After(timeEnd) {
		e := errors.New(fmt.Sprintf("Time start after time end in line %s: %s", data[1], err))
		return DataIn{}, e
	}

	// Write end and start times
	formattingData.timeStart = timeStart //Format("15:04")
	formattingData.timeEnd = timeEnd     //Format("15:04")

	//Write cost for table
	formattingData.costForHour, err = strconv.Atoi(data[2])
	if err != nil || formattingData.costForHour < 0 {
		e := errors.New(fmt.Sprintf("Error with cost for table in line '%s': %s", data[2], err))
		return DataIn{}, e
	}

	r, _ := regexp.Compile("^[a-z0-9-_]+$")

	//Check and write subjects
	for i := 3; i < len(data); i++ {
		subjectString := strings.Split(data[i], " ")

		//Check len of line data
		if len(subjectString) < 3 || len(subjectString) > 4 {
			e := errors.New(fmt.Sprintf("Too much subjects in line: %s", data[i]))
			return DataIn{}, e
		}

		subject := Subject{}

		//Write time of subj
		subject.timeOfSubj, err = time.Parse("15:04", subjectString[0])
		if err != nil || subject.timeOfSubj.After(timeEnd) {
			e := errors.New(fmt.Sprintf("Error in subject with time in line '%s': %s", data[i], err))
			return DataIn{}, e
		}

		//Write subj id
		subject.id, err = strconv.Atoi(subjectString[1])
		if err != nil {
			e := errors.New(fmt.Sprintf("Error with subject id in line '%s': %s", data[i], err))
			return DataIn{}, e
		}

		// check  0<id<5
		if subject.id > 4 || subject.id < 0 {
			e := errors.New(fmt.Sprintf("Unknown id in line '%s'", data[i]))
			return DataIn{}, e
		}

		if len(subjectString) == 4 {
			subject.name = subjectString[2]
			matched := r.MatchString(subject.name)
			if !matched {
				e := errors.New(fmt.Sprintf("Bad name in line '%s'", data[i]))
				return DataIn{}, e
			}
			subject.tableNumb, err = strconv.Atoi(subjectString[3])
			if err != nil || subject.tableNumb > formattingData.tables {
				e := errors.New(fmt.Sprintf("Bad table number in line '%s': %s", data[i], err))
				return DataIn{}, e
			}
		} else {
			subject.name = subjectString[2]
			matched := r.MatchString(subject.name)
			if !matched {
				e := errors.New(fmt.Sprintf("Bad name in line '%s'", data[i]))
				return DataIn{}, e
			}
		}

		formattingData.subjects = append(formattingData.subjects, subject)
	}

	//КОСТЫЛИ КОСТЫЛИ КОСТЫЛИ ЕСЛИ ОСТАНЕТСЯ ВРЕМЯ ПОДУМАТЬ КАК ЗАПИХНУТЬ ЭТО В ПОТОК
	for i := 1; i < len(formattingData.subjects); i++ {
		if formattingData.subjects[i-1].timeOfSubj.After(formattingData.subjects[i].timeOfSubj) {
			e := errors.New(fmt.Sprintf("Bad time in lines: '%s' after '%s'", data[i+2], data[i+3]))
			return DataIn{}, e
		}
	}

	return formattingData, err
}

func formatDataOut(data DataIn) (DataOut, error) {
	dataOut := DataOut{}
	var err error

	dataOut.timeStart = data.timeStart
	dataOut.timeEnd = data.timeEnd

	stackInSubjects := StackSubj[Subject]{}

	// Fill stack of our in subjects
	// fmt.Println("Fill the stack -----------------------")
	for i := len(data.subjects) - 1; i >= 0; i-- {
		stackInSubjects.Push(data.subjects[i])
	}

	// fmt.Println(stackInSubjects, "\nMy stack -------------------------")
	// fmt.Println(len(stackInSubjects.data))
	// tables := map[int]bool{}
	tables := []Tables{}
	for i := 0; i < data.tables; i++ {
		tables = append(tables, Tables{
			owner: "none",
			// number: i + 1,
			// timeInUsed: 0,
			// pay:        10,
		})

	}

	clientsQueue := list.New()

	// tables := map[interface{}]interface{}{}
	clintsOnTables := map[string]int{} // КОСТЫЛИ КОСТЫЛИ
	clientsInClub := map[string]interface{}{}

	// clients := map[string]bool{}

	//Hardcore check all subjects
	for {
		if stackInSubjects.IsEmpty() {
			break
		}

		subjectIn := stackInSubjects.Pop()
		subjectOut := Subject{}

		if data.timeEnd.Before(subjectIn.timeOfSubj) {
			break
		}

		dataOut.subjectsOut = append(dataOut.subjectsOut, subjectIn)
		//Generate errors for clients
		errosOfClients := []string{"NotOpenYet", "YouShallNotPass", "ClientUnknown", "ICanWaitNoLonger!", "PleaseIsBusy"}

		//Not open yet error |||| it could be bettr \-_-/
		if subjectIn.timeOfSubj.Before(data.timeStart) {
			subjectOut.timeOfSubj = subjectIn.timeOfSubj
			subjectOut.id = 13
			subjectOut.name = errosOfClients[0]
			dataOut.subjectsOut = append(dataOut.subjectsOut, subjectOut)
			continue
		}

		tableClear := false
		for i := 0; i < data.tables; i++ {
			if tables[i].owner == "none" {
				tableClear = true
			}
		}

		//ID1 : Клиент пришел
		if subjectIn.id == 1 {
			//Клиент уже в клубе
			if clientsInClub[subjectIn.name] != nil {
				subjectOut.timeOfSubj = subjectIn.timeOfSubj
				subjectOut.id = 13
				subjectOut.name = errosOfClients[1]
				dataOut.subjectsOut = append(dataOut.subjectsOut, subjectOut)
				continue
			} else {
				clientsInClub[subjectIn.name] = true
				// clintsOnTables[subjectIn.name] = nil
			}
			//ID3 : Клиент ожидает
		} else if subjectIn.id == 3 {
			//Есть свободные столы а клиент ждет
			if tableClear {
				subjectOut.timeOfSubj = subjectIn.timeOfSubj
				subjectOut.id = 13
				subjectOut.name = errosOfClients[3]
				dataOut.subjectsOut = append(dataOut.subjectsOut, subjectOut)
				continue
				//Очередь клиентов больше числа столов
			} else if clientsQueue.Len() > data.tables {
				subjectOut.timeOfSubj = subjectIn.timeOfSubj
				subjectOut.id = 11
				subjectOut.name = subjectIn.name
				dataOut.subjectsOut = append(dataOut.subjectsOut, subjectOut)
				continue
				//Встал в очередь в ожидании компа
			} else if clientsQueue.Len() < data.tables && !tableClear {
				clientsQueue.PushBack(subjectIn.name)
				// fmt.Printf("%s stand in queue in time %s \n", subjectIn.name, subjectIn.timeOfSubj.Format("15:04"))
			}
			//ID2 : Клиент сел за стол
		} else if subjectIn.id == 2 {
			//Неизвестный клиент
			if clientsInClub[subjectIn.name] == nil {
				subjectOut.timeOfSubj = subjectIn.timeOfSubj
				subjectOut.id = 13
				subjectOut.name = errosOfClients[2]
				dataOut.subjectsOut = append(dataOut.subjectsOut, subjectOut)
				continue
			}

			//Пытается сесть за занятый стол
			if tables[subjectIn.tableNumb-1].owner != "none" {
				subjectOut.timeOfSubj = subjectIn.timeOfSubj
				subjectOut.id = 13
				subjectOut.name = errosOfClients[4]
				dataOut.subjectsOut = append(dataOut.subjectsOut, subjectOut)
				continue
			} else {
				//Освобождаем стол от клиента если он пересаживается и оплачивание стола
				if clintsOnTables[subjectIn.name] != 0 {
					tables[clintsOnTables[subjectIn.name]-1].owner = "none"
					owned := subjectIn.timeOfSubj.Sub(tables[clintsOnTables[subjectIn.name]-1].startOwned)
					// fmt.Println("в обращении", owned)
					tables[clintsOnTables[subjectIn.name]-1].timeInUsed += owned
					needToPay := 0
					if owned.Minutes() > 0 {
						hours := int(owned.Hours() + 1)
						needToPay = hours * data.costForHour
					} else {
						hours := int(owned.Hours())
						needToPay = hours * data.costForHour
					}
					tables[clintsOnTables[subjectIn.name]-1].pay += needToPay
					tables[clintsOnTables[subjectIn.name]-1].endOwned = subjectIn.timeOfSubj
				}
				// Сажаем за стол
				tables[subjectIn.tableNumb-1].owner = subjectIn.name
				tables[subjectIn.tableNumb-1].startOwned = subjectIn.timeOfSubj
				clintsOnTables[subjectIn.name] = subjectIn.tableNumb
			}
		} else if subjectIn.id == 4 {
			// Такого клиента нет
			if clientsInClub[subjectIn.name] == nil {
				subjectOut.timeOfSubj = subjectIn.timeOfSubj
				subjectOut.id = 13
				subjectOut.name = errosOfClients[2]
				dataOut.subjectsOut = append(dataOut.subjectsOut, subjectOut)
				continue
			} else {
				clientsInClub[subjectIn.name] = nil

				// Сажаем за пк чела из очереди если она есть
				if clintsOnTables[subjectIn.name] != 0 {
					//Собираем дань!!!
					owned := subjectIn.timeOfSubj.Sub(tables[clintsOnTables[subjectIn.name]-1].startOwned)
					// fmt.Println("Owned:", owned, subjectIn.timeOfSubj.Format("15:04"), tables[clintsOnTables[subjectIn.name]-1].startOwned.Format("15:04"))
					tables[clintsOnTables[subjectIn.name]-1].timeInUsed += owned
					needToPay := 0
					if owned.Minutes() > 0 {
						hours := int(owned.Hours() + 1)
						// fmt.Println(hours, owned)
						needToPay = hours * data.costForHour
					} else {
						hours := int(owned.Hours())
						// fmt.Println(hours)
						needToPay = hours * data.costForHour
					}
					// fmt.Printf("Клиент %s уходит и должен заплатить %v\n", subjectIn.name, needToPay)
					tables[clintsOnTables[subjectIn.name]-1].pay += needToPay
					tables[clintsOnTables[subjectIn.name]-1].endOwned = subjectIn.timeOfSubj
					if clientsQueue.Len() > 0 {
						client := clientsQueue.Front().Value
						clientsQueue.Remove(clientsQueue.Front())
						var name string

						if str, ok := client.(string); ok {
							name = str
						}
						// fmt.Println(name)
						subjectOut.timeOfSubj = subjectIn.timeOfSubj
						subjectOut.id = 12
						subjectOut.name = name
						subjectOut.tableNumb = clintsOnTables[subjectIn.name]
						dataOut.subjectsOut = append(dataOut.subjectsOut, subjectOut)
						clintsOnTables[name] = clintsOnTables[subjectIn.name]
						tables[clintsOnTables[subjectIn.name]-1].owner = name
						tables[clintsOnTables[subjectIn.name]-1].startOwned = subjectIn.timeOfSubj
						continue
					}
					tables[clintsOnTables[subjectIn.name]-1].owner = "none"
					clintsOnTables[subjectIn.name] = 0
					continue
				}
			}
		}
	}

	for i := range tables {
		if tables[i].owner != "none" {
			owned := data.timeEnd.Sub(tables[i].startOwned)
			tables[i].timeInUsed += owned
			needToPay := 0
			if owned.Minutes() > 0 {
				hours := int(owned.Hours() + 1)
				needToPay = hours * data.costForHour
			} else {
				hours := int(owned.Hours())
				needToPay = hours * data.costForHour
			}
			tables[i].pay += needToPay
		}
		re := regexp.MustCompile("[0-9]+")
		mainTime := re.FindAllString(tables[i].timeInUsed.String(), -1)
		// fmt.Println(mainTime)
		if len(mainTime) > 1 {
			if len(mainTime[1]) < 2 {
				mainTime[1] = "0" + mainTime[1]
			}
		}

		workingTime := "00:00"
		if len(mainTime) > 1 {
			workingTime = mainTime[0] + ":" + mainTime[1]
		}

		dataOut.payment = append(dataOut.payment, Pay{
			table:       i + 1,
			money:       tables[i].pay,
			workingTime: workingTime,
		})
	}

	sort.SliceStable(tables, func(i, j int) bool {
		return tables[i].owner < tables[j].owner
	})

	for _, v := range tables {
		if v.owner != "none" {
			dataOut.subjectsOut = append(dataOut.subjectsOut, Subject{
				timeOfSubj: data.timeEnd,
				id:         11,
				name:       v.owner,
			})
		}
	}

	return dataOut, err
}

func outputData(data DataOut) {
	fmt.Println(data.timeStart.Format("15:04"))

	for _, s := range data.subjectsOut {
		l := ""
		if s.tableNumb != 0 {
			l = strconv.Itoa(s.tableNumb)
		}
		fmt.Println(s.timeOfSubj.Format("15:04"), s.id, s.name, l)
	}

	fmt.Println(data.timeEnd.Format("15:04"))

	for _, s := range data.payment {
		fmt.Println(s.table, s.money, s.workingTime)
	}
}

func main() {
	fmt.Println("Processing(+): Read data")
	if len(os.Args) < 2 {
		log.Fatalf("Missing parameter, provide file name!")
	}

	data, err := readData(os.Args[1])
	if err != nil {
		log.Fatalf("Error with read data: %s", err)
	}

	fmt.Println("Processing(+): Fill the data in good format")
	dataIn, err := formatDataIn(data)
	if err != nil {
		log.Fatalf("Error with formatting data input: %s", err)
	}

	fmt.Println("Processing(+): Formatting data")
	dataOut, err := formatDataOut(dataIn)
	if err != nil {
		log.Fatalf("Error with formatting data out: %s", err)
	}

	outputData(dataOut)
}
