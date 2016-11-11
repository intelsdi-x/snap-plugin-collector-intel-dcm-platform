package ipmi

import (
	"fmt"
	"strings"
)

type FruParser struct {
	IpmiLayer IpmiAL
}

var FruProcessor = &FruParser{}

var CmdFruHeader = IpmiRequest{[]byte{0xa, 0x10, 0x0}, 0x0, 0x0}
var CmdFruData = IpmiRequest{[]byte{0xa, 0x11, 0x0, 0x0, 0x0, 0x0}, 0x0, 0x0}
var CmdDeviceId = IpmiRequest{[]byte{0x6, 0x1}, 0x0, 0x0}
var CmdBMCMac = IpmiRequest{[]byte{0xc, 0x2, 0x1, 0x5, 0x0, 0x0}, 0x0, 0x0}

func (fp *FruParser) GetInventoryInfo(host string) (map[string]string, error) {

	// read fru inventory info
	response, err := fp.IpmiLayer.ExecRaw(CmdFruHeader, host)
	if err != nil {
		return nil, err
	}
	var data []byte
	data = response.Data[1:]

	var fruInventoryAreaSize, areaAccessedLength uint16

	fruInventoryAreaSize = uint16(data[1])<<8 + uint16(data[0])

	if fruInventoryAreaSize <= 0 {
		return nil, fmt.Errorf("%d : Invalid fruInventoryAreaSize", fruInventoryAreaSize)
	}
	areaAccessedLength = uint16((data[2] & 0x1) + 1)

	// first read 8 bytes
	CmdFruData.Data[5] = 8
	response, err = fp.IpmiLayer.ExecRaw(CmdFruData, host)
	if err != nil {
		return nil, err
	}

	data = response.Data[1:]
	if data == nil || len(data) <= 8 {
		return nil, err
	}

	//fruChassisInfoAreaStartingOffset = 	(short)(((data[3] & 0xff) * 8));
	//fruBoardInfoAreaStartingOffset = (short)(((data[4] & 0xff) * 8));

	// read product info area length
	fruProductInfoAreaStartingOffset := uint16(((data[5] & 0xff) * 8))

	GenerateFruData(fruProductInfoAreaStartingOffset/areaAccessedLength, 2)

	response, err = fp.IpmiLayer.ExecRaw(CmdFruData, host)
	if err != nil {
		return nil, err
	}

	data = response.Data[1:]
	fruProductInfoAreaLen := uint16((data[2] & 0xff) * 8)

	if fruProductInfoAreaLen == 0 || fruProductInfoAreaStartingOffset+fruProductInfoAreaLen > fruInventoryAreaSize {
		return nil, fmt.Errorf("%d : Invalid fruProductInfoSize", fruProductInfoAreaLen)
	}

	fruProductInfoAreaData := make([]byte, fruProductInfoAreaLen)
	indexInFruProductInfoArea := uint16(0)

	byteLentoRead := uint16(16)
	if fruProductInfoAreaLen <= 16 {
		byteLentoRead = fruProductInfoAreaLen
	}

	for {

		GenerateFruData((fruProductInfoAreaStartingOffset+
			indexInFruProductInfoArea*16)/areaAccessedLength, byteLentoRead)
		response, err = fp.IpmiLayer.ExecRaw(CmdFruData, host)
		if err != nil {
			break
		}
		data = response.Data[1:]

		copy(fruProductInfoAreaData[indexInFruProductInfoArea*16:indexInFruProductInfoArea*16+uint16(data[0])], data[1:])

		if (indexInFruProductInfoArea+1)*16 < fruProductInfoAreaLen {
			byteLentoRead = 16
			if fruProductInfoAreaLen-(indexInFruProductInfoArea+1)*16 <= 16 {
				byteLentoRead = fruProductInfoAreaLen - (indexInFruProductInfoArea+1)*16
			}

			indexInFruProductInfoArea++

		} else {
			break
		}

	}

	var metrics = []string{
		"inventory/product_manufacturer",
		"inventory/product_name",
		"",
		"",
		"inventory/product_serial",
	}
	ret := map[string]string{}
	offset := uint16(3)
	for _, metric := range metrics {
		if offset >= fruProductInfoAreaLen {
			break
		}
		if metric != "" {
			ret[metric] = GetFruAreaString(fruProductInfoAreaData[offset:])
		}
		offset = uint16(fruProductInfoAreaData[offset]&0x3f) + offset + 1
	}

	// get Firmware version
	response, err = fp.IpmiLayer.ExecRaw(CmdDeviceId, host)
	if err == nil {
		// parser Firmware version
		data = response.Data[1:]
		major := data[2] & 0x7f
		minor := ((data[3]&0xf0)>>4)*10 + (data[3] & 0x0f)
		ret["inventory/firmware_version"] = fmt.Sprintf("%d.%d", major, minor)
	}

	// get BMC MAC
	response, err = fp.IpmiLayer.ExecRaw(CmdBMCMac, host)
	if err == nil {
		data = response.Data[1:]
		ret["inventory/bmc_mac"] = fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", data[1], data[2], data[3], data[4], data[5], data[6])
	}

	return ret, nil
}

func GenerateFruData(offset uint16, length uint16) {
	CmdFruData.Data[3] = byte(offset & 0xff)
	CmdFruData.Data[4] = byte((offset >> 8) & 0xff)
	CmdFruData.Data[5] = byte(length)
}

func GetFruAreaString(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	offset := uint16(0)

	bcd := []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		' ', '-', '.', ':', ',', '_'}

	typeCode := ((data[offset] & 0xC0) >> 6)
	length := uint16(data[offset])
	offset++
	length &= 0x3f

	if offset+length-1 >= uint16(len(data)) {
		return ""
	}

	//trim the end mark
	var index uint16
	for index = 0; index < length; index++ {
		if (data[offset+index] == 0) || (data[offset+index] == 0xFF) {
			break
		}
	}
	length = uint16(index)

	if length == 0 {
		return ""
	}

	switch typeCode {
	case 0:
	case 3:
		buf := make([]byte, length)
		for i := 0; i < len(buf); i++ {
			buf[i] = data[offset+uint16(i)]
		}
		return strings.TrimSpace(string(buf[:]))
	case 2:
		var chars [4]byte
		var i, j uint16
		buf1 := make([]byte, ((((length + 2) * 4) / 3) &^ 3))
		for i = 0; i < length; i += 3 {
			j = 0

			//k = ((len - i) < 3 ? (len-i) : 3);
			k := uint16(3)
			if length-i < 3 {
				k = length - i
			}
			value := data[offset+i]
			// chars[3] = value > 0 ? value : (value + 256)
			chars[3] = value

			//value = (k > 1 ? data[offset + i + 1] : 0)
			value = 0
			if k > 1 {
				value = data[offset+i+1]
			}

			// chars[2] = value > 0 ? value : (value + 256)
			chars[2] = value

			//value = (k > 2 ? data[offset + i + 2] : 0)
			value = 0
			if k > 2 {
				value = data[offset+i+2]
			}

			//chars[1] = value > 0 ? value : (value + 256)
			chars[1] = value

			for k = 0; k < 4; k++ {
				buf1[j+1] = (chars[3] & 0x3f) + 0x20
				j++

				m := 3
				for ; m > 0; m-- {
					chars[m] = (chars[m] >> 6) |
						((chars[m-1] & 0x3f) << 2)
				}
				chars[m] = chars[m] >> 6
			}
		}
		return strings.TrimSpace(string(buf1[:]))

	case 1:
		buf1 := make([]byte, length)
		var j uint16
		for j = 0; j < length; j++ {
			buf1[j] = bcd[(data[offset+j] & 0x0f)]
		}
		return strings.TrimSpace(string(buf1[:]))
	}

	return ""
}
