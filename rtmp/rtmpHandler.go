package rtmp

import (
	"rtmpServerStudy/amf"

	"encoding/hex"
	"fmt"
	"github.com/nareix/bits/pio"
)

func (self *Session) writeDataMsg(csid, msgsid uint32, args ...interface{}) (err error) {
	return self.writeAMF0Msg(RtmpMsgAmfCMD, csid, msgsid, args...)
}

func (self *Session) writeCommandMsg(csid, msgsid uint32, args ...interface{}) (err error) {
	return self.writeAMF0Msg(RtmpMsgAmfCMD, csid, msgsid, args...)
}

func (self *Session) DoSend(b []byte, csid uint32, timestamp uint32, msgtypeid uint8, msgsid uint32, msgdatalen int) (n int, err error) {

	pos := 0
	sn := 0
	last := self.writeMaxChunkSize
	end := msgdatalen
	testn := 0
	for msgdatalen > 0 {
		if pos == 0 {
			n := self.fillChunk0Header(self.chunkHeaderBuf, csid, timestamp, msgtypeid, msgsid, msgdatalen)
			fmt.Print(hex.Dump(self.chunkHeaderBuf[:n]))
			testn, err = self.bufw.Write(self.chunkHeaderBuf[:n])
			fmt.Printf("1-----------:%d\n", testn)
		} else {
			n := self.fillChunk3Header(self.chunkHeaderBuf, csid, timestamp)
			fmt.Print(hex.Dump(self.chunkHeaderBuf[:n]))
			testn, err = self.bufw.Write(self.chunkHeaderBuf[:n])
			fmt.Printf("2-----------:%d\n", testn)
		}
		if msgdatalen > self.writeMaxChunkSize {
			fmt.Printf("3*************:pos:%d****************last:%d\n", pos, last)
			fmt.Print(hex.Dump(b[pos:last]))
			if sn, err = self.bufw.Write(b[pos:last]); err != nil {
				return
			}

			pos += sn
			last += sn
			msgdatalen -= sn
			continue
		}
		fmt.Print(hex.Dump(b[pos:end]))
		fmt.Printf("4************:pos:%d****************end:%d\n", pos, end)
		if sn, err = self.bufw.Write(b[pos:end]); err != nil {
			return
		}
		pos += sn
		msgdatalen -= sn
		return
	}
	return
}

func (self *Session) writeAMF0Msg(msgtypeid uint8, csid, msgsid uint32, args ...interface{}) (err error) {

	size := 0
	for _, arg := range args {
		size += amf.LenAMF0Val(arg)
	}
	b := self.GetWriteBuf(size)
	n := 0

	for _, arg := range args {
		n += amf.FillAMF0Val(b[n:], arg)
	}
	fmt.Println("========================")
	fmt.Print(hex.Dump(b[:n]))
	fmt.Println("========================")

	_, err = self.DoSend(b, csid, 0, msgtypeid, msgsid, size)
	return
}

func (self *Session) writeBasicConf() (err error) {
	// > SetChunkSize
	if err = self.writeSetChunkSize(self.writeMaxChunkSize); err != nil {
		return
	}
	// > WindowAckSize
	if err = self.writeWindowAckSize(5000000); err != nil {
		return
	}
	// > SetPeerBandwidth

	if err = self.writeSetPeerBandwidth(5000000, 2); err != nil {
		return
	}
	return
}

func (self *Session) writeStreamBegin(msgsid uint32) (err error) {
	b := self.GetWriteBuf(chunkHeaderLength + 6)
	n := self.fillChunk0Header(b, 2, 0, RtmpMsgUser, 0, 6)
	pio.PutU16BE(b[n:], RtmpUserStreamBegin)
	n += 2
	pio.PutU32BE(b[n:], msgsid)
	n += 4
	_, err = self.bufw.Write(b[:n])
	return
}
