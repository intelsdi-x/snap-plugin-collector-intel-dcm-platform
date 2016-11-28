package ipmi

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
)

type SdrParser struct {
	IpmiLayer IpmiAL
}

type ComponentHealth struct {
	Type string
	Severity string
	SensorInfos []SensorInfo
}

type SensorInfo struct {
	Name string	
	ErrorCode uint16    				
	ErrorDescription string
	Severity string
}

type DeviceId struct{
	IsDeviceSdr bool
}

type SdrHeader struct {
	NextRecordId uint16
	RecordId     uint16
	SdrVersion   uint16
	RecordType   uint16
	RecordLength uint16
}

type SdrInfo struct {
	Header      SdrHeader
	SensorNumber uint16
	SensorType uint16
	EventReadingType uint16
}

type SensorStatus struct {
	SensorNumber   uint16
	Status uint16
	StateUnavailable bool
	SensorType uint16
	ReadingType uint16
}

var CmdGetDeviceId = IpmiRequest{[]byte{0x6, 0x1}, 0x0, 0x0}

var CmdStorageSdrInfo = IpmiRequest{[]byte{0xa, 0x20}, 0x0, 0x0}
var CmdDeviceSdrInfo = IpmiRequest{[]byte{0x4, 0x20}, 0x0, 0x0}

var CmdReserveStorageSdr = IpmiRequest{[]byte{0xa, 0x22}, 0x0, 0x0}
var CmdReserveDeviceSdr = IpmiRequest{[]byte{0x4, 0x22}, 0x0, 0x0}

var CmdGetSdrRepositoryAllocationInfo = IpmiRequest{[]byte{0xa, 0x21}, 0x0, 0x0}

//AddData byte[6]: 
//byte[0] =reservationId[0]
//byte[1] = reservationId[1]
//byte[2,3] = recordId
//byte[4] = offsetIntoRecord
//byte[5] = bytesToRead
var CmdGetStorageSdr = IpmiRequest{[]byte{0xa, 0x23,0x0,0x0,0x0,0x0,0x0,0x0}, 0x0, 0x0}
var CmdGetDeviceSdr = IpmiRequest{[]byte{0x4, 0x21,0x0,0x0,0x0,0x0,0x0,0x0}, 0x0, 0x0}

var CmdGetSensorReading = IpmiRequest{[]byte{0x4, 0x2D,0x0}, 0x0, 0x0}

type SensorHealthSetting struct{
	SensorType uint16
	ReadingType uint16
	Status []uint16
}

var SensorHealthConfig = map[string]SensorHealthSetting{
	"1:1":SensorHealthSetting{1,1,[]uint16{0,1,2,3,4,5,6,7,8,9,10,11,12}},
	"1:3":SensorHealthSetting{1,3,[]uint16{1}},
	"1:5":SensorHealthSetting{1,5,[]uint16{1}},
	"2:1":SensorHealthSetting{2,1,[]uint16{0,1,2,3,4,5,6,7,8,9,10,11,12}},
	"2:3":SensorHealthSetting{2,3,[]uint16{1}},
	"4:1":SensorHealthSetting{4,1,[]uint16{0,1,2,3,4,5,6,7,8,9,10,11,12}},
	"4:3":SensorHealthSetting{4,3,[]uint16{1}},
	"4:7":SensorHealthSetting{4,7,[]uint16{2,3,5}},
	"4:11":SensorHealthSetting{4,11,[]uint16{1,5}},
	"7:1":SensorHealthSetting{7,1,[]uint16{0,1,2,3,4,5,6,7,8,9,10,11,12}},
	"7:3":SensorHealthSetting{7,3,[]uint16{1}},
	"7:111":SensorHealthSetting{7,111,[]uint16{0,1,2,3,4,5,6,11}},
	"8:3":SensorHealthSetting{8,3,[]uint16{1}},
	"8:7":SensorHealthSetting{8,7,[]uint16{2,3,5}},
	"8:11":SensorHealthSetting{8,11,[]uint16{1,5}},
	"8:111":SensorHealthSetting{8,111,[]uint16{1,2,3,4,5,6}},
	"12:1":SensorHealthSetting{12,1,[]uint16{0,1,2,3,4,5,6,7,8,9,10,11,12}},
	"12:111":SensorHealthSetting{12,111,[]uint16{1,3,7,10}},
	"12:3":SensorHealthSetting{12,3,[]uint16{1}},
	"13:111":SensorHealthSetting{13,111,[]uint16{1,2,5,6,8}},
	"13:3":SensorHealthSetting{13,3,[]uint16{1}},
	"41:111":SensorHealthSetting{41,111,[]uint16{0,1}}}

type ComponentDescription struct{
	ComponentType string
	Metrics string
}

//SensorType componentType mapping
var SensorTypeComponentMap = map[uint16]ComponentDescription{
	1:ComponentDescription{"TEMPERATURE_HEALTH","health/temperature"},
	2:ComponentDescription{"VOLTAGE_HEALTH","health/voltage"},
	4:ComponentDescription{"FAN_HEALTH","health/fan"},
	7:ComponentDescription{"PROCESSOR_HEALTH","health/processor"},
	8:ComponentDescription{"POWERSUPPLY_HEALTH","health/powersupply"},
	12:ComponentDescription{"MEMORY_HEALTH","health/memory"},
	13:ComponentDescription{"STORAGE_HEALTH","health/storage"},
	41:ComponentDescription{"BATTERY_HEALTH","health/battery"}}

func (sp *SdrParser) GetComponentHealth(host string)(map[string]string,error){
	ret := map[string]string{}
	deviceId,err := sp.GetDeviceId(host)
	if err != nil{
		return nil,err
	}
	sdrInfos,err := sp.ScanSdr(deviceId.IsDeviceSdr,host)
	if err != nil{
		return nil,err
	}

	sdrStatus,err := sp.GetSdrData(sdrInfos,host)
	if err != nil{
		return nil,err
	}

	for _,sdrStat := range sdrStatus{
		if sdrStat.StateUnavailable {
			continue
		}
	
		var componentType ComponentDescription= SensorTypeComponentMap[sdrStat.SensorType]
		if &componentType != nil{
			//threshold sensor
			if sdrStat.ReadingType == 1{
				var sensorInfo = GetSensorInfo(sdrStat.Status)
				ret[componentType.Metrics] =  sensorInfo.Severity
			}else{
				var configKey = fmt.Sprintf("%v:%v", sdrStat.SensorType, sdrStat.ReadingType)
				var sensorHealthSetting =SensorHealthConfig[configKey]
				if &sensorHealthSetting != nil{
					for unexpectedStat := range sensorHealthSetting.Status {
						var offset uint16= 1 << uint16(unexpectedStat)
						if (sdrStat.Status & offset) > 0 {
							ret[componentType.Metrics] = "SEV_UNKNOWN"
						}
					}
				}
			}
		}

	}

	return ret,nil
}

func (sp *SdrParser) GetDeviceId(host string) (*DeviceId, error) {
	var deviceId DeviceId
	response, err := sp.IpmiLayer.ExecRaw(CmdGetDeviceId, host)
	if err != nil {
	    log.WithFields(log.Fields{
			"error": err,
		}).Debug("GetDeviceId error")		
		return nil, err
	}
	if len(response.Data) < 1{
		return nil, fmt.Errorf("Unexpected response data: No add data in response")
	}
	var data = response.Data[1:]

	deviceId.IsDeviceSdr =  false;		

	if((data[1] & 0x80) != 0){
		if((data[5] & 0x2) != 0){
			//use SDR repository
			deviceId.IsDeviceSdr = false;
		}else{
			if((data[5] & 0x1) != 0){
				deviceId.IsDeviceSdr = true;
			}else{
					return nil, fmt.Errorf("Fail to find a valuable SDR")				
				}
			}
		}else{
			deviceId.IsDeviceSdr = false;
		}

    return &deviceId,nil
}

func (sp *SdrParser) GetSdrInfo(isDeviceSdr bool,host string) (error) {
    var cmd IpmiRequest
    if (!isDeviceSdr){
        cmd = CmdStorageSdrInfo
    }else {
        cmd = CmdDeviceSdrInfo
    }
	_,err := sp.IpmiLayer.ExecRaw(cmd, host)
	return err
}

// return reservation Id
func (sp *SdrParser) ReserveSdr(isDeviceSdr bool,host string) ([]byte,error) {
	var cmd IpmiRequest
    if (!isDeviceSdr){
        cmd = CmdReserveStorageSdr
    }else {
        cmd = CmdReserveDeviceSdr
    }	
	response, err := sp.IpmiLayer.ExecRaw(cmd, host)
	if (err != nil){
		return nil, err
	}
	if len(response.Data) < 1{
		return nil, fmt.Errorf("Unexpected response data: No add data in response")
	}	
	var data = response.Data[1:]
	return data,nil
}

func (sp *SdrParser) GetSdrRepositoryAllocationInfo(host string) (uint16, error) {
	var cmd IpmiRequest = CmdGetSdrRepositoryAllocationInfo
	response, err := sp.IpmiLayer.ExecRaw(cmd, host)
	if err != nil {
		return uint16(0), err
	}
	if len(response.Data) < 1{
		return uint16(0), fmt.Errorf("Unexpected response data: No add data in response")
	}
    var data = response.Data[1:]
	if len(data) != 9 {			
		return uint16(0),fmt.Errorf("Additional data result unexpected!");
	}
	var allocationUnitSize = uint16(data[3])<< 8+ uint16(data[2])
	var sdrMaxReadLen uint16= 16
	if allocationUnitSize != 0 {
		sdrMaxReadLen = allocationUnitSize;
	}
	return sdrMaxReadLen,nil	
}

func (sp *SdrParser) GetSdr(reservationId []byte,recordId uint16, offsetIntoRecord uint16, byetsToRead uint16, isDeviceSdr bool,host string)([]byte,error){
    var cmd IpmiRequest
    if !isDeviceSdr{
        cmd = CmdGetStorageSdr.Clone()
    }else {
        cmd = CmdGetDeviceSdr.Clone()
    }
	//AddData
	if reservationId != nil{
		cmd.Data[2] = reservationId[0]
		cmd.Data[3] = reservationId[1]
	}

	cmd.Data[4] = byte (recordId)
	cmd.Data[5] = byte (recordId >> 8)
	cmd.Data[6] = byte (offsetIntoRecord)
	cmd.Data[7] = byte (byetsToRead)
	response, err := sp.IpmiLayer.ExecRaw(cmd, host)
	if err != nil{
		return nil, err
	}
	if len(response.Data) < 1{
		return nil, fmt.Errorf("Unexpected response data: No add data in response")
	}		
	var data = response.Data[1:]
	return data,nil
}

func (sp *SdrParser) GetSensorReading(sensorNum int,host string)([]byte,error){
	var cmd IpmiRequest = CmdGetSensorReading.Clone()
	cmd.Data[2] = byte(sensorNum) 
	response, err := sp.IpmiLayer.ExecRaw(cmd, host)
	if (err != nil){
		return nil, err
	}
	if len(response.Data) < 1{
		return nil, fmt.Errorf("Unexpected response data: No add data in response")
	}	
	var data = response.Data[1:]
	return data,nil
}

func (sp *SdrParser) ScanSdr(isDeviceSdr bool,host string) ([]SdrInfo, error){
	err := sp.GetSdrInfo(isDeviceSdr,host)
	if (err != nil){		
		return nil,err
	}
	reservationId,err := sp.ReserveSdr(isDeviceSdr,host)
	if err != nil{
		log.WithFields(log.Fields{
				"host": host,
				"isDeviceSdr":isDeviceSdr,
				"error":err,
		}).Debug("ScanSdr ReserveSdr exit with error")			
		return nil,err
	}
/*	log.WithFields(log.Fields{
			"host": host,
			"isDeviceSdr":isDeviceSdr,
			"reservationId":reservationId,
		}).Debug("ScanSdr reservationId")	*/
	var sdrMaxReadLen = uint16(16)
	if !isDeviceSdr {
		sdrMaxRead,err := sp.GetSdrRepositoryAllocationInfo(host)
		if (err != nil){
			log.WithFields(log.Fields{
					"host": host,
					"isDeviceSdr":isDeviceSdr,
					"error":err,
			}).Debug("ScanSdr GetSdrRepositoryAllocationInfo exit with error")				
			return nil,err
		}
		sdrMaxReadLen = sdrMaxRead
	}	
/*	log.WithFields(log.Fields{
			"reservationId":reservationId,
			"host": host,
			"isDeviceSdr":isDeviceSdr,
			"sdrMaxRead":sdrMaxReadLen,
		}).Debug("ScanSdr sdrMaxRead")		*/
	return sp.ScanSdrLoop(reservationId,0,sdrMaxReadLen,isDeviceSdr,host)
}


func (sp *SdrParser) ScanSdrLoop(reservationId []byte,recordId uint16,sdrMaxReadLen uint16,isDeviceSdr bool,host string) ([]SdrInfo, error){
	var recId  = recordId
	var sdrSet []SdrInfo = []SdrInfo{}
	log.WithFields(log.Fields{
		"host": host,
		"isDeviceSdr":isDeviceSdr,
		"reservationId":reservationId,
		"recordId":recordId,
		"sdrMaxReadLen":sdrMaxReadLen,
		}).Debug("ScanSdrLoop")
	for {
		//reservationId []byte，recordId int,isDeviceSdr bool,host string
		  sdrHeader,err:=sp.GetSdrHeader(reservationId,recId,isDeviceSdr,host)
		  if err !=nil{	
			  return nil,err
		  }
/*		 log.WithFields(log.Fields{
			"host": host,
			"isDeviceSdr":isDeviceSdr,
			"reservationId":reservationId,
			"recordId":recordId,
			"sdrMaxReadLen":sdrMaxReadLen,
			"sdrHeader.NextRecordId":sdrHeader.NextRecordId,		
			"sdrHeader.RecordType":sdrHeader.RecordType,
			"sdrHeader.RecordId":sdrHeader.RecordId,
			"sdrHeader.RecordLength":sdrHeader.RecordLength,
			}).Debug("ScanSdrLoop GetSdrHeader ")	*/		  
		  var recordType  = sdrHeader.RecordType
		  if (recordId == 0 || sdrHeader.RecordId == recordId )&& 
			  //IpmiConstants.SDR_TYPE_FUL,IpmiConstants.SDR_TYPE_COMPACT
			  (  recordType == 0x01 || recordType == 0x02){	
				sdr,err := sp.GetSdrByHeader(reservationId,sdrHeader,sdrMaxReadLen,isDeviceSdr,host)
				if err !=nil{
					return nil,err
				}
/*				log.WithFields(log.Fields{
					"host": host,
					"isDeviceSdr":isDeviceSdr,
					"reservationId":reservationId,
					"recordId":recordId,
					"sdrMaxReadLen":sdrMaxReadLen,
					"sdr.Header.NextRecordId":sdr.Header.NextRecordId,		
					"sdr.Header.RecordType":sdr.Header.RecordType,
					"sdr.Header.RecordId":sdr.Header.RecordId,
					"sdr.Header.RecordLength":sdr.Header.RecordLength,
					"sdr.EventReadingType":sdr.EventReadingType,
					"sdr.SensorNumber":sdr.SensorNumber,
					"sdr.SensorType":sdr.SensorType,
				}).Debug("ScanSdrLoop GetSdrByHeader ")	*/				
				recordType = sdr.Header.RecordType
				if recordType == uint16 (0x01) || recordType == uint16 (0x02) {
					sdrSet = append(sdrSet,sdr)
				  }										
		 }
		 recId = sdrHeader.NextRecordId 
		 if recId == uint16(0xFFFF){
			 break
		 }	
	}
	return sdrSet,nil
}

func(sp * SdrParser) GetSdrBytes(reservationId []byte, recordId uint16,totalBytesToRead uint16, sdrMaxReadLen uint16,isDeviceSdr bool,host string)([]byte, error){
	var data []byte = make([]byte, totalBytesToRead + uint16(2))
	var bytesLeftToRead = totalBytesToRead
	var currentBytesToRead uint16= 0	
	for bytesRead := uint16(0);  bytesRead < bytesLeftToRead; bytesRead += currentBytesToRead {					
		bytesLeftToRead = uint16 (totalBytesToRead - bytesRead)
		if (bytesLeftToRead > sdrMaxReadLen) {
			currentBytesToRead = sdrMaxReadLen;
		} else {
			currentBytesToRead = bytesLeftToRead;
	    }			
		var currentSdr []byte
		currentSdr,err := sp.GetSdr(reservationId,recordId,bytesRead,currentBytesToRead,isDeviceSdr,host)
		if err !=nil{
			return nil,err
		}		
		// First copy the next-record-id data
		data[0] = currentSdr[0];
		data[1] = currentSdr[1];	
		// Then the rest of the data
		for i := uint16(0);i < currentBytesToRead; i++ {
			data[uint16(2) + bytesRead + i] = currentSdr[uint16(2) + i]		
		}		
	} 
/*	log.WithFields(log.Fields{
		"reservationId": reservationId,
		"recordId":recordId,
		"totalBytesToRead":totalBytesToRead,
		"sdrMaxReadLen":sdrMaxReadLen,
		"isDeviceSdr":isDeviceSdr,
		"host":host,
		"data":data,
	}).Debug("GetSdrBytes ")	*/
	return data,nil
}

func (sp *SdrParser) GetSdrHeader(reservationId []byte,recordId uint16,isDeviceSdr bool,host string) (*SdrHeader, error){
	var sdrHeaderByte []byte
	var header *SdrHeader = &SdrHeader{}
	sdrHeaderByte,err := sp.GetSdrBytes(reservationId,recordId,5,16,isDeviceSdr,host)
	if err!=nil{
		return nil,err
	}
	//NextRecordId 0-2
	header.NextRecordId = uint16 (sdrHeaderByte[0]) | uint16 (sdrHeaderByte[1]) << 8
	header.RecordId  = uint16 (sdrHeaderByte[2]) | uint16 (sdrHeaderByte[3]) << 8
	header.SdrVersion = uint16 (sdrHeaderByte[4])
	header.RecordType = uint16 (sdrHeaderByte[5])
	header.RecordLength = uint16 (sdrHeaderByte[6]) 
	return header,nil
}

func (sp *SdrParser) GetSdrByHeader(reservationId []byte,header *SdrHeader,sdrMaxReadLen uint16,isDeviceSdr bool,host string) (SdrInfo, error){
	var sdrInfo SdrInfo
	var totalBytesToRead = header.RecordLength + uint16 (5)
	sdrBytes,err :=	sp.GetSdrBytes(reservationId,header.RecordId,totalBytesToRead, sdrMaxReadLen,isDeviceSdr,host)
	if err!=nil{
		return sdrInfo,err
	}

	//Header: 6 bytes
	sdrInfo.Header.NextRecordId =  uint16 (sdrBytes[0]) | uint16 (sdrBytes[1]) << 8
	sdrInfo.Header.RecordId  = uint16 (sdrBytes[2]) | uint16 (sdrBytes[3]) << 8
	sdrInfo.Header.SdrVersion = uint16 (sdrBytes[4])
	sdrInfo.Header.RecordType = uint16 (sdrBytes[5]) 
	sdrInfo.Header.RecordLength = uint16 (sdrBytes[6])

	//SDR_TYPE_FULL,SDR_TYPE_COMPACT
	if sdrInfo.Header.RecordType == uint16 (0x01) ||sdrInfo.Header.RecordType == uint16(0x02){
		sdrInfo.SensorNumber = uint16 (sdrBytes[9]) 
		sdrInfo.SensorType = uint16 (sdrBytes[14]) 
		sdrInfo.EventReadingType = uint16 (sdrBytes[15]) 		
	}else {
		return sdrInfo,fmt.Errorf("Unexpected RecordType")
	}
	return sdrInfo,nil
}

func (sp *SdrParser) GetSdrData(sdrs []SdrInfo,host string) ([]SensorStatus, error){
	var sensorStatus []SensorStatus
	sensorStatus = make([]SensorStatus,len(sdrs))
	var cmd = CmdGetSensorReading.Clone()
	for i,sdr := range sdrs{
		cmd.Data[3] = byte(sdr.SensorNumber)
		response, err := sp.IpmiLayer.ExecRaw(cmd, host)
		if err != nil{
			return nil,err
		}
		if len(response.Data) < 1{
			return nil, fmt.Errorf("Unexpected response data: No add data in response")
		}		
		var data = response.Data[1:]
		sensorStatus[i].StateUnavailable = false
		// data[0] sensor reading
		// data[1]
		// [6] 0b - sensor scanning disabled
		// [5] 1b - reading/state unavailable
		// [4:0] reserved, Ignore on read
		if ((data[1] & 0x20) != 0 || (data[1] & 0x40) == 0) {					
			sensorStatus[i].StateUnavailable = true
			continue
		}
		var readingType = sdr.EventReadingType
		var status uint16
		if (readingType == 0x6F || (readingType >= 0x02 && readingType <= 0x0C)) {
			// discrete SDR
			status = uint16(data[2]);
			if (len(data) >= 4) {
				status = status | ((uint16 (data[3]) & 0x7F) << 8);
			}				
		}else{
			if (readingType == 0x01) {
				// threshold SDR
				status = uint16 (data[2]) & 0x3F;
			}
		}
		sensorStatus[i].ReadingType = sdr.EventReadingType
		sensorStatus[i].SensorNumber =sdr.SensorNumber
		sensorStatus[i].SensorType = sdr.SensorType
		sensorStatus[i].Status = status
	} 
	return sensorStatus,nil
	
}

func GetSensorInfo (reading uint16) SensorInfo{
	var ret SensorInfo

	if (reading&(1<<5)) != 0 {
		ret.ErrorDescription = "at or above upper non-recoverable threshold"
		ret.ErrorCode = uint16(0x06)
		ret.Severity = "CRITICAL"
	}else if (reading&(1<<4)) != 0 {
		ret.ErrorDescription = "at or above upper critical threshold"
		ret.ErrorCode = uint16(0x05)
		ret.Severity = "CRITICAL"
	}else if (reading&(1<<3)) != 0 {
		ret.ErrorDescription =  "at or above upper non-critical threshold"
		ret.ErrorCode = uint16(0x04)
		ret.Severity = "WARNING"
	}else if (reading&(1<<2)) != 0 {
		ret.ErrorDescription = "at or below lower non-recoverable threshold"
		ret.ErrorCode = uint16(0x03)
		ret.Severity = "CRITICAL"
	}else if (reading&(1<<1)) != 0{
		ret.ErrorDescription = "at or below lower critical threshold"
		ret.ErrorCode = uint16(0x02)
		ret.Severity = "CRITICAL"
	}else if (reading&(1<<0)) != 0 {
		ret.ErrorDescription =  "at or below lower non-critical threshold"
		ret.ErrorCode = uint16(0x01)
		ret.Severity = "WARNING"
	}else{
		ret.ErrorDescription = ""
		ret.ErrorCode = uint16(0x00)
		ret.Severity = "OK"
	}
	return ret
}





