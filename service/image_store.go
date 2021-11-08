package service

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"os"
	"sync"
)

// ImageStore is a interface to store laptop image
type ImageStore interface {
	// Save saves a new laptop image to the store
	Save(laptopID string, imageType string, imageData bytes.Buffer) (string, error)
}

// DiskImageStore stores image on the disk and its info on memory
type DiskImageStore struct {
	mutex sync.RWMutex
	imageFolder string
	images map[string]*ImageInfo
}

// ImageInfo contains information of laptop image
type ImageInfo struct {
	LaptopID string
	Type string
	Path string
}

// NewDiskImageStore return a new DiskImageStore
// 生成一个硬盘存储图片的对象
func NewDiskImageStore(imageFolder string) *DiskImageStore {
	return &DiskImageStore{
		imageFolder: imageFolder,
		images: make(map[string]*ImageInfo),
	}
}

// Save save laptop image to disk and keep info to memory
func (store *DiskImageStore) Save(laptopID string, imageType string, imageData bytes.Buffer) (string, error) {
	// 生成uuid，作为image的名称
	imageId, err := uuid.NewUUID()
	if err != nil {
		return "", fmt.Errorf("cannot generate image id: %w", err)
	}
	// 构造image存储路径
	imagePath := fmt.Sprintf("%s/%s%s", store.imageFolder, imageId, imageType)

	// 创建image文件
	file, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("cannot create image file: %w", err)
	}

	// 将上传过来的image保存到创建的文件中
	_, err = imageData.WriteTo(file)
	if err != nil {
		return "", fmt.Errorf("cannot write image file: %w", err)
	}

	// image保存成功后，将info存到内存信息中 map
	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.images[imageId.String()] = &ImageInfo{
		LaptopID: laptopID,
		Type: imageType,
		Path: imagePath,
	}
	// 返回imageID给上传image的用户
	return imageId.String(), nil
}






