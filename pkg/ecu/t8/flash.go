package t8

import (
	"bytes"
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"time"

	"github.com/roffe/txlogger/pkg/ecu/t8legion"
	"github.com/roffe/txlogger/pkg/ecu/t8util"
)

func (t *Client) FlashECU(ctx context.Context, bin []byte) error {
	if len(bin) != 0x100000 {
		return errors.New("err: Invalid T8 file size")
	}

	if err := t.legion.Bootstrap(ctx, false); err != nil {
		return err
	}

	t.cfg.OnMessage("Comparing MD5's for erase")
	t.cfg.OnProgress(-9)
	t.cfg.OnProgress(0)
	for i := 1; i <= 9; i++ {
		lmd5 := t8util.GetPartitionMD5(bin, 6, i)
		md5, err := t.legion.GetMD5(ctx, t8legion.GetTrionic8MD5, uint16(i))
		if err != nil {
			return err
		}
		t.cfg.OnMessage(fmt.Sprintf("local partition   %d> %X", i, lmd5))
		t.cfg.OnMessage(fmt.Sprintf("remote partitioncfg  %d> %X", i, md5))
		t.cfg.OnProgress(float64(i))
	}
	start := time.Now()
	err := t.legion.WriteFlash(ctx, t8legion.EcuByte_T8, 0x100000, bin, false)
	if err != nil {
		return err
	}

	t.cfg.OnMessage("Verifying md5..")

	ecuMD5bytes, err := t.legion.IDemand(ctx, t8legion.GetTrionic8MD5, 0x00)
	if err != nil {
		return err
	}
	calculatedMD5 := md5.Sum(bin)

	t.cfg.OnMessage(fmt.Sprintf("Remote MD5 : %X", ecuMD5bytes))
	t.cfg.OnMessage(fmt.Sprintf("Local MD5  : %X", calculatedMD5))

	if !bytes.Equal(ecuMD5bytes, calculatedMD5[:]) {
		return errors.New("md5 Verification failed")
	}

	t.cfg.OnMessage("Done, took: " + time.Since(start).String())

	return nil
}
