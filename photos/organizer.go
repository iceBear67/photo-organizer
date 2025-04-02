package photos

import (
	"github.com/oriser/regroup"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type PhotoOrganizer struct {
	checkSize bool
	RootDir   string
	Rules     []*regroup.ReGroup
}

func CreateOrganizer(checkSize bool, rootDir string) *PhotoOrganizer {
	return &PhotoOrganizer{
		checkSize: checkSize,
		RootDir:   rootDir,
		Rules:     make([]*regroup.ReGroup, 0),
	}
}

type Ops map[string]*PhotoInfo

func (o *PhotoOrganizer) Prepare(overwrite bool, photo []*PhotoInfo, dstDir string) *Ops {
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
		fs, err := os.Stat(path)
		shouldWrite := false
		if os.IsNotExist(err) {
			shouldWrite = true
		} else {
			if ele.Size != -1 {
				if ele.Size != fs.Size() {
					log.Printf("!!! Size mismatch: %v, %v (%v -> %v)", ele.SourceFile, path, ele.Size, fs.Size())
					shouldWrite = true
				}
			}
			shouldWrite = shouldWrite || overwrite
		}
		if shouldWrite {
			if _, ok := result[path]; ok {
				log.Printf("CONFLICT FILE: %v -> %v ", ele.SourceFile, path)
				hasError = true
			} else {
				result[path] = ele
			}
		}
	}
	if hasError {
		return nil
	}
	return (*Ops)(&result)
}

func (o *PhotoOrganizer) CollectInfo() []*PhotoInfo {
	result := make([]*PhotoInfo, 0)
	o.traverseDirectory(o.RootDir, &result)
	checkDates(result)
	log.Printf("Collected %v photos", len(result))
	return result
}

func checkDates(pResult []*PhotoInfo) {
	yearMedium := -1
	result := pResult
	for _, pFile := range result {
		file := *pFile
		if file.Date != nil && len(file.Date) >= 1 {
			if yearMedium == -1 {
				yearMedium = file.Date[0]
			}
			if math.Abs(float64(yearMedium-file.Date[0])) > 100 {
				log.Println("Found images that crosses ~100 years: ", file.FileName)
			}
		}
	}
}

func (o *PhotoOrganizer) traverseDirectory(dir string, result *[]*PhotoInfo) {
	log.Println("Traversing ", dir)
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("Failed to enumerate files in target directory %v. %v", dir, err)
		return
	}
	for _, file := range files {
		entryPath := filepath.Join(dir, file.Name())
		if file.IsDir() {
			o.traverseDirectory(entryPath, result)
			continue
		}
		pi := o.toPhotoInfo(entryPath)
		if o.checkSize {
			var fi os.FileInfo
			if fi, err = file.Info(); err != nil {
				log.Printf("Failed to retrieve file info of %v, skipping this for size check!", entryPath)
			} else {
				pi.Size = fi.Size()
			}
		}
		if !pi.isComplete() {
			log.Printf("Incomplete photo info for file %v: %v", file.Name(), pi.SourceFile)
		}
		*result = append(*result, pi)
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
