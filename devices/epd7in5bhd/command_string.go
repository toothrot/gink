// Code generated by "stringer -type=command"; DO NOT EDIT.

package epd7in5bhd

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[setGateDriver-1]
	_ = x[setGateDrivingVoltage-3]
	_ = x[setSourceDrivingVoltage-4]
	_ = x[deepSleepMode-16]
	_ = x[dataEntryMode-17]
	_ = x[displayRefresh-18]
	_ = x[hvReadyDetection-20]
	_ = x[vciDetection-21]
	_ = x[tempSensorControl-24]
	_ = x[tempSensorWrite-26]
	_ = x[tempSensorRead-26]
	_ = x[tempSensorControlExt-28]
	_ = x[masterActivation-32]
	_ = x[displayUpdateControl1-33]
	_ = x[displayUpdateControl2-34]
	_ = x[writeRAMBW-36]
	_ = x[writeRAMRed-38]
	_ = x[readRAM-39]
	_ = x[vcomSense-40]
	_ = x[vcomSenseDuration-41]
	_ = x[vcomOTP-42]
	_ = x[vcomControlRegister-43]
	_ = x[vcomWriteRegister-44]
	_ = x[otpRegisterRead-45]
	_ = x[crcCalculation-52]
	_ = x[crcStatusRead-53]
	_ = x[otpProgramSelect-54]
	_ = x[displayOptionRegister-55]
	_ = x[userOptionRegister-56]
	_ = x[borderWaveformControl-60]
	_ = x[readRamOption-65]
	_ = x[setRamXStart-68]
	_ = x[setRamYStart-69]
	_ = x[autoWriteRamRed-70]
	_ = x[autoWriteRamBW-71]
	_ = x[setRamXAddressCtr-78]
	_ = x[setRamYAddressCtr-79]
	_ = x[softStart-12]
}

const _command_name = "setGateDriversetGateDrivingVoltagesetSourceDrivingVoltagesoftStartdeepSleepModedataEntryModedisplayRefreshhvReadyDetectionvciDetectiontempSensorControltempSensorWritetempSensorControlExtmasterActivationdisplayUpdateControl1displayUpdateControl2writeRAMBWwriteRAMRedreadRAMvcomSensevcomSenseDurationvcomOTPvcomControlRegistervcomWriteRegisterotpRegisterReadcrcCalculationcrcStatusReadotpProgramSelectdisplayOptionRegisteruserOptionRegisterborderWaveformControlreadRamOptionsetRamXStartsetRamYStartautoWriteRamRedautoWriteRamBWsetRamXAddressCtrsetRamYAddressCtr"

var _command_map = map[command]string{
	1:  _command_name[0:13],
	3:  _command_name[13:34],
	4:  _command_name[34:57],
	12: _command_name[57:66],
	16: _command_name[66:79],
	17: _command_name[79:92],
	18: _command_name[92:106],
	20: _command_name[106:122],
	21: _command_name[122:134],
	24: _command_name[134:151],
	26: _command_name[151:166],
	28: _command_name[166:186],
	32: _command_name[186:202],
	33: _command_name[202:223],
	34: _command_name[223:244],
	36: _command_name[244:254],
	38: _command_name[254:265],
	39: _command_name[265:272],
	40: _command_name[272:281],
	41: _command_name[281:298],
	42: _command_name[298:305],
	43: _command_name[305:324],
	44: _command_name[324:341],
	45: _command_name[341:356],
	52: _command_name[356:370],
	53: _command_name[370:383],
	54: _command_name[383:399],
	55: _command_name[399:420],
	56: _command_name[420:438],
	60: _command_name[438:459],
	65: _command_name[459:472],
	68: _command_name[472:484],
	69: _command_name[484:496],
	70: _command_name[496:511],
	71: _command_name[511:525],
	78: _command_name[525:542],
	79: _command_name[542:559],
}

func (i command) String() string {
	if str, ok := _command_map[i]; ok {
		return str
	}
	return "command(" + strconv.FormatInt(int64(i), 10) + ")"
}
