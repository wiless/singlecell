package main

import (
	"fmt"
	"log"
	"math"

	"github.com/spf13/viper"
	"github.com/wiless/vlib"
)

//AppConfig  Struct for the app parameteres
type AppConfig struct {
	ISD              float64
	TxPowerDbm       float64
	Out2IndoorLossDb float64

	UENoiseFigureDb float64
	BSNoiseFigureDb float64
	INDOORRatio     float64
	INCARRatio      float64
	INCARLossdB     float64
	ActiveCells     int
	TrueCells       int // The number of cells where the UEs are dropped ..
	AntennaVTilt    float64
	Extended        bool
	ForceAllLOS     bool
	ShadowLoss      bool
	BandwidthMHz    float64
	BSHeight        float64
	UEHeight        float64
	LogInfo         bool
	NumUEperCell    int
}

var C AppConfig // Global Configuration variable

func (C *AppConfig) SetDefaults() {
	C.INDOORRatio = 0
	C.INCARRatio = 0
	C.INCARLossdB = 0
	C.Out2IndoorLossDb = 0
	C.ActiveCells = -1 // Default all the cells are active
	C.TrueCells = -1
	C.Extended = false
	C.ForceAllLOS = false
	C.BandwidthMHz = 10
	C.UENoiseFigureDb = 7
	C.BSNoiseFigureDb = 5
	C.ShadowLoss = true
	C.LogInfo = false
	C.NumUEperCell = 30
	// C.TrueCells = -1   // Default to all the cells
	// Do for others too
}

// ReadAppConfig reads all the configuration for the app
func ReadAppConfig() {
	C.SetDefaults()
	log.Print("Reading APP config ")
	viper.AddConfigPath(indir)
	viper.SetConfigName("config")

	err := viper.ReadInConfig()
	if err != nil {
		log.Print("ReadInConfig ", err)
	}
	// Set all the default values
	{
		viper.SetDefault("TxPowerDbm", TxPowerDbm)
		viper.SetDefault("ISD", ISD)
		viper.SetDefault("INDOORRatio", C.INDOORRatio)
		viper.SetDefault("INCARRatio", C.INCARRatio)
		viper.SetDefault("INCARLossdB", C.INCARLossdB)
		viper.SetDefault("Out2IndoorLossDb", C.Out2IndoorLossDb)
		viper.SetDefault("UENoiseFigureDb", C.UENoiseFigureDb)
		viper.SetDefault("BSNoiseFigureDb", C.BSNoiseFigureDb)

		viper.SetDefault("ActiveCells", C.ActiveCells)
		viper.SetDefault("TrueCells", C.TrueCells)
		viper.SetDefault("ForceAllLOS", C.ForceAllLOS)
		CellRadius = ISD / math.Sqrt(3.0)
		log.Print(C)
	}
	err = viper.Unmarshal(&C)
	if err == nil {
		log.Print("Error unmarshalling ", err)
	}
	log.Printf("%#v ", C)
	// Load from the external configuration files
	ISD = viper.GetFloat64("ISD")
	TxPowerDbm = viper.GetFloat64("TxpowerDBm")

	CellRadius = ISD / math.Sqrt(3.0)
	fmt.Print(C)

	SaveAppConfig()

}

func SaveAppConfig() {
	SwitchOutput()
	vlib.SaveStructure(C, "OutputSetting.json", true)
	SwitchBack()

}
