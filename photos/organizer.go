package photos

import (
	"github.com/oriser/regroup"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type PhotoOrganizer struct {
	events  chan *PhotoInfo
	RootDir string
	Rules   []*regroup.ReGroup
}

func CreateOrganizer(rootDir string) *PhotoOrganizer {
	return &PhotoOrganizer{
		RootDir: rootDir,
		Rules:   make([]*regroup.ReGroup, 0),
		events:  make(chan *PhotoInfo, 16),
	}
}

type Ops map[string]*PhotoInfo

func (o *PhotoOrganizer) Prepare(photo []*PhotoInfo, dstDir string) *Ops {
	result := make(map[string]*PhotoInfo, len(photo))
	log.Printf("Preparing for %v photos", len(photo))
	hasError := false
	for _, ele := range photo {
		path := filepath.Join(dstDir)
		dateComponents := len(ele.Date)
		if dateComponents >= 1 {
			path = filepath.Join(path, strconv.Itoa(ele.Date[0]))
			if dateComponents >= 2 {
				path = filepath.Join(path, strconv.Itoa(ele.Date[1]))
			}
		}
		path = filepath.Join(path, ele.FileName)
		if _, ok := result[path]; ok {
			log.Printf("CONFLICT FILE: %v -> %v ", ele.SourceFile, path)
			hasError = true
		} else {
			result[path] = ele
		}
	}
	if hasError {
		return nil
	}
	return (*Ops)(&result)
}

func (o *PhotoOrganizer) CollectInfo() []*PhotoInfo {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go o.traverseDirectory(o.RootDir, &wg, true)
	go func() {
		time.Sleep(100 * time.Millisecond)
		wg.Wait()
		close(o.events)
	}()
	result := o.selectLoop()
	log.Printf("Collected %v photos", len(*result))
	return *result
}

func (o *PhotoOrganizer) selectLoop() *[]*PhotoInfo {
	result := make([]*PhotoInfo, 0)
	yearMedium := -1
	for file := range o.events {
		if file.Date != nil && len(file.Date) >= 1 {
			if yearMedium == -1 {
				yearMedium = file.Date[0]
			}
			if math.Abs(float64(yearMedium-file.Date[0])) > 100 {
				log.Println("Found images that crosses ~100 years: ", file.FileName)
			}
		}
		result = append(result, file)
	}
	return &result
}

func (o *PhotoOrganizer) traverseDirectory(dir string, wg *sync.WaitGroup, first bool) {
	log.Println("Traversing ", dir)
	if first {
		defer wg.Add(-1)
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("Failed to enumerate files in target directory %v. %v", dir, err)
		return
	}
	for _, file := range files {
		entryPath := filepath.Join(dir, file.Name())
		if file.IsDir() {
			o.traverseDirectory(entryPath, wg, false)
			continue
		}
		pi := o.toPhotoInfo(entryPath)
		if !pi.isComplete() {
			log.Printf("Incomplete photo info for file %v: %v", file.Name(), pi.SourceFile)
		}
		o.events <- pi
	}
}

func (o *PhotoOrganizer) toPhotoInfo(file string) *PhotoInfo {
	var mostComplete *PhotoInfo
	maxFoundComponentsNum := 0
	for _, e := range o.Rules {
		pi := createPhotoInfo(file)
		groups, err := e.Groups(pi.FileName)
		if err != nil {
			continue
		}
		ymd := make([]int, 0)
		if val, ok := groups["Timestamp"]; ok {
			if i, err := strconv.ParseInt(val, 10, 64); err == nil {
				t := time.UnixMilli(i)
				y, m, d := t.Date()
				ymd = append(append(append(ymd, y), int(m)), d)
			}
		} else {
			if yr, err := strconv.Atoi(groups["Year"]); err == nil {
				ymd = append(ymd, yr)
				if mn, err := strconv.Atoi(groups["Month"]); err == nil {
					ymd = append(ymd, mn)
					if day, err := strconv.Atoi(groups["Day"]); err == nil {
						ymd = append(ymd, day)
					}
				}
			}
		}
		pi.Date = ymd

		current := calcCompleteComponents(pi)
		if current > maxFoundComponentsNum {
			maxFoundComponentsNum = current
			mostComplete = pi
		}
		if pi.isComplete() {
			return pi
		}
	}
	if mostComplete == nil {
		mostComplete = createPhotoInfo(file)
	}
	if maxFoundComponentsNum == 0 {
		log.Println("Error extracting Y/M/D info from ", mostComplete.FileName)
	}
	return mostComplete
}
