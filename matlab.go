package main

import "github.com/wiless/vlib"

type MatInfo struct {
	BaseID             int `json:"baseID"`
	SecID              int `json:"secID"`
	UserID             int `json:"userID"`
	RSSI               float64
	IfStation          vlib.VectorI `json:"ifStation"`
	IfRSSI             vlib.VectorF `json:"ifRSSI"`
	ThermalNoise       float64      `json:"thermalNoise"`
	SINR               float64
	RestOfInterference float64 `json:"restOfInterference"`
	Distance           float64 `json:"Distance"`
}

var matlab *vlib.Matlab
