package photos

import (
	"log"
	"path/filepath"
)

type PhotoInfo struct {
	Size       int64
	Date       []int
	FileName   string
	SourceFile string
}

func createPhotoInfo(source string) *PhotoInfo {
	base := filepath.Base(source)
	if base == "" {
		log.Println("WARNING: photo file name is empty: ", source)
	}
	return &PhotoInfo{
		-1, make([]int, 0), base, source,
	}
}

func (p *PhotoInfo) isComplete() bool {
	return calcCompleteComponents(p) == 3
}

func calcCompleteComponents(pi *PhotoInfo) int {
	if pi.Date == nil {
		return 0
	}
	return len(pi.Date)
}
