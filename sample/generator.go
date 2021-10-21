package sample

import (
	"github.com/golang/protobuf/ptypes"
	"pc_book/pd"
)

// NewKeyboard return a new sample keyboard
func NewKeyboard() *pd.Keyboard {
	keyboard := &pd.Keyboard{
		Layout: randomKeyboardLayout(),
		Backlit: randomKeyboardBacklit(),
	}
	return keyboard
}

// NewCPU return a new sample CPU
func NewCPU() *pd.CPU {
	brand := randomCPUBrand()
	name := randomCPUName(brand)

	coreNumber := randomInt(2, 8)
	threadNumber := randomInt(coreNumber, 12)

	minGhz := randomFloat64(2.0, 3.5)
	maxGhz := randomFloat64(minGhz, 5.0)

	return &pd.CPU{
		Brand: brand,
		Name: name,
		NumberCores: uint32(coreNumber),
		NumberThreads: uint32(threadNumber),
		MinGhz: minGhz,
		MaxGhz: maxGhz,
	}
}

// NewGPU return a new sample CPU
func NewGPU() *pd.GPU {
	brand := randomGPUBrand()
	name := randomGPUName(brand)

	minGhz := randomFloat64(2.0, 3.5)
	maxGhz := randomFloat64(minGhz, 5.0)

	memory := &pd.Memory{
		Value: uint64(randomInt(2, 6)),
		Unit: pd.Memory_GIGABYTE,
	}

	return &pd.GPU{
		Brand: brand,
		Name: name,
		MinGhz: minGhz,
		MaxGhz: maxGhz,
		Memory: memory,
	}
}

// NewSSD returns a new sample SSD
func NewSSD() *pd.Storage {
	memGB := randomInt(128, 1024)

	ssd := &pd.Storage{
		Driver: pd.Storage_SSD,
		Memory: &pd.Memory{
			Value: uint64(memGB),
			Unit:  pd.Memory_GIGABYTE,
		},
	}
	return ssd
}

// NewHDD returns a new sample HDD
func NewHDD() *pd.Storage {
	memTB := randomInt(1, 6)

	hdd := &pd.Storage{
		Driver: pd.Storage_HDD,
		Memory: &pd.Memory{
			Value: uint64(memTB),
			Unit:  pd.Memory_TERABYTE,
		},
	}
	return hdd
}

// NewRAM returns a new sample RAM
func NewRAM() *pd.Memory {
	memGB := randomInt(4, 64)

	ram := &pd.Memory{
		Value: uint64(memGB),
		Unit:  pd.Memory_GIGABYTE,
	}

	return ram
}

// NewScreen returns a new sample Screen
func NewScreen() *pd.Screen {
	screen := &pd.Screen{
		SizeInch:   randomFloat32(13, 17),
		Resolution: randomScreenResolution(),
		Panel:      randomScreenPanel(),
		Multitouch: randomBool(),
	}

	return screen
}

// NewLaptop returns a new sample Laptop
func NewLaptop() *pd.Laptop {
	brand := randomLaptopBrand()
	name := randomLaptopName(brand)

	laptop := &pd.Laptop{
		Id:       randomID(),
		Brand:    brand,
		Name:     name,
		Cpu:      NewCPU(),
		Ram:      NewRAM(),
		Gpus:     []*pd.GPU{NewGPU()},
		Storages: []*pd.Storage{NewSSD(), NewHDD()},
		Screen:   NewScreen(),
		Keyboard: NewKeyboard(),
		Weight: &pd.Laptop_WeightKg{
			WeightKg: randomFloat64(1.0, 3.0),
		},
		PriceUsd:    randomFloat64(1500, 3500),
		ReleaseYear: uint32(randomInt(2015, 2019)),
		UpdatedAt:   ptypes.TimestampNow(),
	}

	return laptop
}

// RandomLaptopScore returns a random laptop score
func RandomLaptopScore() float64 {
	return float64(randomInt(1, 10))
}



