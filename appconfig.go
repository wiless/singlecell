package main

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"math"

	"github.com/spf13/viper"
	"github.com/wiless/vlib"
)

//AppConfig  Struct for the app parameteres
type AppConfig struct {
	CarriersGHz      float64
	ISD              float64
	TxPowerDbm       float64
	Out2IndoorLossDb float64
	UENoiseFigureDb  float64
	BSNoiseFigureDb  float64
	INDOORRatio      float64
	INCARRatio       float64
	INCARLossdB      float64
	// ActiveCells     int
	NCells        int
	ActiveBSCells int // The number of cells where the BS are enabled (for link)
	ActiveUECells int // The number of cells where the UEs are dropped ..
	AntennaVTilt  float64
	Extended      bool
	ForceAllLOS   bool
	ShadowLoss    bool
	BandwidthMHz  float64
	BSHeight      float64
	UEHeight      float64
	LogInfo       bool
	NumUEperCell  int
	UEcells       []int
	BScells       []int
}

// C Global Configuration variable
var C AppConfig

// SetDefaults loads the default values for the simulation
func (C *AppConfig) SetDefaults() {
	C.CarriersGHz = .7 // 700MHz
	C.INDOORRatio = 0
	C.INCARRatio = 0
	C.INCARLossdB = 0
	C.Out2IndoorLossDb = 0
	C.NCells = 19
	C.ActiveBSCells = -1 // Default all the cells are active
	C.ActiveUECells = -1 // UEs are dropped in all the cells
	C.Extended = false
	C.ForceAllLOS = false
	C.BandwidthMHz = 10
	C.UENoiseFigureDb = 7
	C.BSNoiseFigureDb = 5
	C.ShadowLoss = true
	C.LogInfo = false
	C.NumUEperCell = 30
	C.UEcells = []int{0, 10}
	C.BScells = []int{0, 1, 2}
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

	fmt.Printf("\n INPUT CONFIGURATION %#v", C)
	err = viper.Unmarshal(&C)
	if err != nil {
		log.Print("Error unmarshalling ", err)
	}

	if C.ActiveBSCells == -1 {
		if len(C.BScells) > 0 {
			C.ActiveBSCells = len(C.BScells)
		} else {
			C.ActiveBSCells = C.NCells
			C.BScells = vlib.NewSegmentI(0, C.ActiveBSCells)
		}
	} else {
		C.BScells = vlib.NewSegmentI(0, C.ActiveBSCells)
	}

	if C.ActiveUECells == -1 {
		if len(C.UEcells) > 0 {
			C.ActiveUECells = len(C.UEcells)
		} else {
			C.ActiveUECells = C.NCells
			C.UEcells = vlib.NewSegmentI(0, C.ActiveUECells)
		}
	} else {
		C.UEcells = vlib.NewSegmentI(0, C.ActiveUECells)
		fmt.Println(C.UEcells)
	}

	// Set all the default values
	// {
	// 	viper.SetDefault("TxPowerDbm", TxPowerDbm)
	// 	viper.SetDefault("ISD", ISD)
	// 	viper.SetDefault("INDOORRatio", C.INDOORRatio)
	// 	viper.SetDefault("INCARRatio", C.INCARRatio)
	// 	viper.SetDefault("INCARLossdB", C.INCARLossdB)
	// 	viper.SetDefault("Out2IndoorLossDb", C.Out2IndoorLossDb)
	// 	viper.SetDefault("UENoiseFigureDb", C.UENoiseFigureDb)
	// 	viper.SetDefault("BSNoiseFigureDb", C.BSNoiseFigureDb)
	// 	viper.SetDefault("ActiveUECells", C.ActiveUECells)
	// 	viper.SetDefault("ActiveBSCells", C.ActiveBSCells)
	// 	viper.SetDefault("ForceAllLOS", C.ForceAllLOS)
	// 	CellRadius = ISD / math.Sqrt(3.0)
	// 	log.Println("AppConfig : ", C)
	// }

	// Load from the external configuration files
	ISD = viper.GetFloat64("ISD")
	TxPowerDbm = viper.GetFloat64("TxpowerDBm")
	CellRadius = ISD / math.Sqrt(3.0)
	CarriersGHz = []float64{C.CarriersGHz}
	SaveAppConfig()

}

func SaveAppConfig() {
	log.Printf("AppConfig : %#v ", C)
	SwitchOutput()
	vlib.SaveStructure(C, "OutputSetting.json", true)
	SwitchBack()

}
