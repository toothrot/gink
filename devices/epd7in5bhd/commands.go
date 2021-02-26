package epd7in5bhd

// See golang.org/x/tools/cmd/stringer
//go:generate stringer -type=command
type command byte

const (
	setGateDriver           command = 0x01
	setGateDrivingVoltage   command = 0x03
	setSourceDrivingVoltage command = 0x04
	deepSleepMode           command = 0x10
	dataEntryMode           command = 0x11
	displayRefresh          command = 0x12
	hvReadyDetection        command = 0x14
	vciDetection            command = 0x15
	tempSensorControl       command = 0x18
	tempSensorWrite         command = 0x1A
	tempSensorRead          command = 0x1A
	tempSensorControlExt    command = 0x1C
	masterActivation        command = 0x20
	displayUpdateControl1   command = 0x21
	displayUpdateControl2   command = 0x22
	writeRAMBW              command = 0x24
	writeRAMRed             command = 0x26
	readRAM                 command = 0x27
	vcomSense               command = 0x28
	vcomSenseDuration       command = 0x29
	vcomOTP                 command = 0x2A
	vcomControlRegister     command = 0x2B
	vcomWriteRegister       command = 0x2C
	otpRegisterRead         command = 0x2D
	crcCalculation          command = 0x34
	crcStatusRead           command = 0x35
	otpProgramSelect        command = 0x36
	displayOptionRegister   command = 0x37
	userOptionRegister      command = 0x38
	borderWaveformControl   command = 0x3C
	readRamOption           command = 0x41
	setRamXStart            command = 0x44
	setRamYStart            command = 0x45
	autoWriteRamRed         command = 0x46
	autoWriteRamBW          command = 0x47
	setRamXAddressCtr       command = 0x4E
	setRamYAddressCtr       command = 0x4F
	softStart               command = 0x0C
)
