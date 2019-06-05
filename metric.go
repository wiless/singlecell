package main

import (
	"encoding/json"
	"fmt"
	"math"

	cell "github.com/wiless/cellular"
	"github.com/wiless/vlib"
)

func EvaluateDIP(RxMetrics map[int]cell.LinkMetric, rxids vlib.VectorI, MAXINTER int, DONORM bool) []MatInfo {

	MatlabResult := make([]MatInfo, len(rxids))

	for indx, rxid := range rxids {
		metric := RxMetrics[rxid]
		var minfo MatInfo
		minfo.UserID = metric.RxNodeID
		minfo.SecID = int(math.Floor(float64(metric.BestRSRPNode) / float64(C.ActiveBSCells)))
		minfo.BaseID = metric.BestRSRPNode

		if metric.TxNodeIDs.Size() < MAXINTER {
			MAXINTER = len(metric.TxNodeIDs) - 1
		}
		// log.Println("METRIC TxNodes ", metric.TxNodeIDs)
		minfo.IfStation = metric.TxNodeIDs[1 : MAXINTER+1] // the first entry is best
		var ifrssi vlib.VectorF
		ifrssi = metric.TxNodesRSRP[1:]

		if DONORM {
			minfo.RSSI = 0 // normalized
			ifrssi = ifrssi.Sub(metric.TxNodesRSRP[0])
		} else {
			minfo.RSSI = metric.TxNodesRSRP[0]

		}

		residual := ifrssi[MAXINTER:]
		residual = vlib.InvDbF(residual)
		ifrssi = ifrssi[0:MAXINTER]

		minfo.IfRSSI = ifrssi // the first entry is best
		minfo.ThermalNoise = metric.N0
		if DONORM {
			minfo.ThermalNoise -= metric.RSSI
		}
		minfo.SINR = metric.BestSINR
		if vlib.Sum(residual) > 0 {
			minfo.RestOfInterference = vlib.Db(vlib.Sum(residual))
		} else {
			minfo.RestOfInterference = -999999
		}

		src := singlecell.Nodes[minfo.BaseID].Location
		dest := singlecell.Nodes[minfo.UserID].Location
		dist := src.DistanceFrom(dest)
		minfo.Distance = dist
		//// test Inf
		_, err := json.Marshal(minfo)
		if err != nil {
			fmt.Printf("\nError %v \n UID %#v , ROI %v", err, minfo, vlib.Sum(residual))
		}
		MatlabResult[indx] = minfo
	}
	return MatlabResult

}
