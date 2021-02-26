package epd7in5bhd

const (
	setGateDriver           byte = 0x01
	setGateDrivingVoltage   byte = 0x03
	setSourceDrivingVoltage byte = 0x04
	deepSleepMode           byte = 0x10
	dataEntryMode           byte = 0x11
	displayRefresh          byte = 0x12
	hvReadyDetection        byte = 0x14
	vciDetection            byte = 0x15
	tempSensorControl       byte = 0x18
	tempSensorWrite         byte = 0x1A
	tempSensorRead          byte = 0x1A
	tempSensorControlExt    byte = 0x1C
	masterActivation        byte = 0x20
	displayUpdateControl1   byte = 0x21
	displayUpdateControl2   byte = 0x22
	writeRAMBW              byte = 0x24
	writeRAMRed             byte = 0x26
	readRAM                 byte = 0x27
	vcomSense               byte = 0x28
	vcomSenseDuration       byte = 0x29
	vcomOTP                 byte = 0x2A
	vcomControlRegister     byte = 0x2B
	vcomWriteRegister       byte = 0x2C
	otpRegisterRead         byte = 0x2D
	crcCalculation          byte = 0x34
	crcStatusRead           byte = 0x35
	otpProgramSelect        byte = 0x36
	displayOptionRegister   byte = 0x37
	userOptionRegister      byte = 0x38
	borderWaveformControl   byte = 0x3C
	readRamOption           byte = 0x41
	setRamXStart            byte = 0x44
	setRamYStart            byte = 0x45
	autoWriteRamRed         byte = 0x46
	autoWriteRamBW          byte = 0x47
	setRamXAddressCtr       byte = 0x4E
	setRamYAddressCtr       byte = 0x4F
	softStart               byte = 0x0C
)
