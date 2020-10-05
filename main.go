package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"sync"
)

type Automobilis struct {
	Make         string  `json:"Make"`
	Year         int     `json:"Year"`
	Displacement float64 `json:"Displacement"`
	Hash         string
}
type Automobiliai struct {
	Auto   []Automobilis
	count  int
	MaxLen int
	mutex  *sync.Mutex
	cond   *sync.Cond
	end    bool
}

func (a *Automobiliai) Insert(aut Automobilis) {
	a.mutex.Lock()
	for a.count == a.MaxLen {
		a.cond.Wait()
	}

	a.Auto[a.count] = aut
	a.count++
	a.cond.Broadcast()
	defer a.mutex.Unlock()
}
func (a *Automobiliai) Remove() Automobilis {
	a.mutex.Lock()
	for a.count == 0 {
		a.cond.Wait()
	}
	result := a.Auto[a.count-1]
	var tmp Automobilis
	a.Auto[a.count-1] = tmp

	a.count--
	a.cond.Broadcast()
	defer a.mutex.Unlock()
	return result
}
func (a *Automobiliai) InsertSort(aut Automobilis) {
	a.mutex.Lock()
	for a.count == a.MaxLen {
		a.cond.Wait()
	}
	j := 0
	for i := 0; i < a.count; i++ {
		if a.Auto[i].Displacement < aut.Displacement {
			j = i
		}
	}
	if j != 0 {
		for i := a.count; i >= j; i-- {
			a.Auto[i+1] = a.Auto[i]
		}
	}
	a.Auto[j] = aut
	a.count++
	a.cond.Broadcast()
	defer a.mutex.Unlock()
}

var Auto []Automobilis

func main() {
	CurrentWD, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	path := CurrentWD + "\\Auto.json"
	jsonFile, err := os.Open(path)

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal(byteValue, &Auto)

	var mutex = sync.Mutex{}
	var cond = sync.NewCond(&mutex)
	MaxLength := 10
	var A = make([]Automobilis, 10)
	Auto1 := Automobiliai{count: 0, MaxLen: MaxLength, cond: cond, mutex: &mutex, Auto: A}

	var B = make([]Automobilis, 40)
	Auto2 := Automobiliai{count: 0, MaxLen: 40, cond: cond, mutex: &mutex, Auto: B}

	var waitGroup = sync.WaitGroup{}
	waitGroup.Add(2)
	for i := 0; i < 2; i++ {
		go execute(&Auto1, &waitGroup, &Auto2)
	}
	for _, s := range Auto {
		Auto1.Insert(s)
	}
	Auto1.end = true
	waitGroup.Wait()
	var RLoc = CurrentWD + "\\Results.txt"
	f, err := os.Create(RLoc)
	defer f.Close()
	f.WriteString(fmt.Sprintf("%15s|%4s|%12s|%50s \n", "Make", "Year", "Displacement", "Hash"))
	for i := 0; i < Auto2.count-1; i++ {
		f.WriteString(fmt.Sprintf("%15s|%4d|%12.2f|%50s \n", Auto2.Auto[i].Make, Auto2.Auto[i].Year, Auto2.Auto[i].Displacement, Auto2.Auto[i].Hash))
	}

	fmt.Println("Program finished execution")
}
func execute(name *Automobiliai, group *sync.WaitGroup, res *Automobiliai) {
	hash := sha256.New()
	for name.end == false && name.count != 0 {
		automobilis := name.Remove()
		if automobilis.Displacement > 2 {
			var ss string
			ss = automobilis.Make + strconv.Itoa(automobilis.Year) + fmt.Sprint(automobilis.Displacement)
			sum := hash.Sum([]byte(ss))
			for i := 0; i < len(sum); i++ {
				ss += string(sum[i])
			}
			automobilis.Hash = ss
			res.InsertSort(automobilis)
		}
	}
	defer group.Done()
}
