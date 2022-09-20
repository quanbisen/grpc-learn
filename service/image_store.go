package service

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"os"
	"sync"
)

type ImageStore interface {
	Save(laptopId string, imageType string, imageData bytes.Buffer) (string, error)
}

type ImageInfo struct {
	LaptopId string
	Type     string
	Path     string
}

type DiskImageStore struct {
	mutex       sync.RWMutex
	imageFolder string
	images      map[string]*ImageInfo
}

func NewDiskImageStore(imageFolder string) ImageStore {
	return &DiskImageStore{
		imageFolder: imageFolder,
		images:      make(map[string]*ImageInfo, 0),
	}
}

func (d *DiskImageStore) Save(laptopId string, imageType string, imageData bytes.Buffer) (string, error) {
	imageId, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("cannot generate image id: %w", err)
	}

	imagePath := fmt.Sprintf("%s/%s.%s", d.imageFolder, imageId, imageType)
	file, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("cannot create image file: %w", err)
	}

	_, err = imageData.WriteTo(file)
	if err != nil {
		return "", fmt.Errorf("cannot write image to file: %w", err)
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.images[imageId.String()] = &ImageInfo{
		LaptopId: laptopId,
		Type:     imageType,
		Path:     imagePath,
	}

	return imageId.String(), nil
}
