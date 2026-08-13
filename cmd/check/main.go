package main

import (
	"fmt"
	"os"

	"github.com/lavrox/SDK-IPWho-Go"
)

var pass, fail int

func ok(c bool, m string) {
	if c {
		pass++
		fmt.Println("  PASS", m)
	} else {
		fail++
		fmt.Println("  FAIL", m)
	}
}
func s(p *string) string {
	if p == nil {
		return "<nil>"
	}
	return *p
}

func main() {
	key := os.Getenv("IPWHO_API_KEY")
	c, err := ipwho.NewClient(key)
	if err != nil {
		fmt.Println("client err:", err)
		os.Exit(2)
	}

	// 1. lookup
	r, err := c.Lookup("8.8.8.8", nil)
	if err != nil {
		fmt.Println("lookup err:", err)
		os.Exit(2)
	}
	d := r.Data
	gl, tz, fl, cu, cn := d.GeoLocation, d.Timezone, d.Flag, d.Currency, d.Connection
	ok(d.IP == "8.8.8.8", "lookup ip == 8.8.8.8")
	ok(gl.Country != nil && *gl.Country == "United States", "lookup country == United States (got "+s(gl.Country)+")")
	ok(cn.AsnNumber != nil && *cn.AsnNumber == 15169, fmt.Sprintf("lookup asnNumber == 15169 (got %v)", cn.AsnNumber))
	ok(gl.DialCode != nil, "dial_code captured ("+s(gl.DialCode)+")")
	ok(gl.IsInEu != nil, "is_in_eu captured")
	ok(tz.TimeZone != nil, "time_zone captured ("+s(tz.TimeZone)+")")
	ok(fl.FlagIcon != nil, "flag_Icon captured ("+s(fl.FlagIcon)+")")
	ok(fl.FlagUnicode != nil, "flag_unicode captured ("+s(fl.FlagUnicode)+")")
	ok(cu.NamePlural != nil, "name_plural captured ("+s(cu.NamePlural)+")")
	ok(cn.AsnOrg != nil, "asn_org captured ("+s(cn.AsnOrg)+")")
	ok(cn.ConnectionType != nil, "connection_type captured ("+s(cn.ConnectionType)+")")

	// 2. me
	me, err := c.Me(nil)
	ok(err == nil && me.Data != nil && me.Data.IP != "", "me ip captured ("+me.Data.IP+")")

	// 3. bulk
	b, err := c.Bulk([]string{"8.8.8.8", "1.1.1.1"}, nil)
	ok(err == nil && len(b) == 2, fmt.Sprintf("bulk returns 2 (got %d)", len(b)))

	// 4. bad key
	bc, _ := ipwho.NewClient("sk.invalid_test_key")
	_, err = bc.Lookup("8.8.8.8", nil)
	ok(err != nil, fmt.Sprintf("bad key raised error (%v)", err))

	fmt.Printf("\nGO RESULT: %d passed, %d failed\n", pass, fail)
	if fail > 0 {
		os.Exit(1)
	}
}
