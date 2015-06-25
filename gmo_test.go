package gmo_test

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/theplant/gmo"
)

var Client = gmo.New(os.Getenv("SiteID"), os.Getenv("SitePass"), os.Getenv("ShopID"), os.Getenv("ShopPass"), gmo.TestEndpoint)

func init() {
	// Fox bypassing the test endpoint insecure certificate error:
	//     "x509: certificate signed by unknown authority"
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

func TestMemberAPIs(t *testing.T) {
	userID := fmt.Sprintf("%v", time.Now().UnixNano())
	if _, err := Client.RegisterMember(userID, "Mr Tester"); err != nil {
		t.Error(err)
	} else if output, err := Client.SearchMember(userID); err != nil {
		t.Error(err)
	} else if output.MemberID != userID || output.MemberName != "Mr Tester" {
		t.Errorf("User should be registered successfully")
	}

	if _, err := Client.UpdateMember(userID, "Miss Tester"); err != nil {
		t.Error(err)
	} else if output, err := Client.SearchMember(userID); err != nil {
		t.Error(err)
	} else if output.MemberID != userID || output.MemberName != "Miss Tester" {
		t.Errorf("User should be updated successfully")
	}

	if _, err := Client.DeleteMember(userID); err != nil {
		t.Error(err)
	} else if _, err := Client.SearchMember(userID); err == nil {
		t.Errorf("User should be deleted successfully")
	}
}

func TestCardAPIs(t *testing.T) {
	userID := fmt.Sprintf("%v", time.Now().UnixNano())
	Client.RegisterMember(userID, "Mr Tester")

	output, err := Client.SaveCard(userID, "4024007154567043", "1010", "Mr Tester")
	if err != nil {
		t.Error(err)
	}
	seq := output.CardSeq
	if _, err := Client.SearchCard(userID, seq); err != nil {
		t.Error(err)
	}

	if _, err := Client.SaveCard(userID, "4024007154567043", "1010", "Miss Tester"); err != nil {
		t.Error(err)
	}
	if output, err := Client.SearchCard(userID, seq); err != nil {
		t.Error(err)
	} else if output.HolderName != "Miss Tester" {
		t.Error("HolderName = %s; want Miss Tester", output.HolderName)
	}

	if _, err := Client.DeleteCard(userID, seq); err != nil {
		t.Error(err)
	}
	if output, err := Client.SearchCard(userID, seq); err != nil {
		t.Error(err)
	} else if output.DeleteFlag != "1" {
		t.Errorf("DeleteFlag = %s; want 1", output.DeleteFlag)
	}
}

func TestCreditCardOrderAPIs(t *testing.T) {
	userID := fmt.Sprintf("%v", time.Now().UnixNano())
	if _, err := Client.RegisterMember(userID, "Mr Tetser"); err != nil {
		t.Fatal(err)
	}
	savedCard, err := Client.SaveCard(userID, "4111111111111111", "0101", "Mr Tetser")
	if err != nil {
		t.Fatal(err)
	}

	orderID := fmt.Sprintf("%v", time.Now().UnixNano())
	t.Log("Order ID:", orderID)
	entryOutput, err := Client.EntryTran(orderID, "1000", "100")
	if err != nil {
		t.Fatal(err)
	}

	execOutput, err := Client.ExecTran(entryOutput.AccessID, entryOutput.AccessPass, orderID, userID, savedCard.CardSeq, "0101")
	if err != nil {
		t.Error(err)
	} else if execOutput.Approve == "" {
		t.Error("Approve should not be empty")
	}

	if _, err := Client.ChangeTran(entryOutput.AccessID, entryOutput.AccessPass, "1500", "100"); err != nil {
		t.Fatal(err)
	}
	if searchOutput, err := Client.SearchTrade(orderID); err != nil {
		t.Error(err)
	} else if searchOutput.Amount != "1500" {
		t.Error("Amount = %s; want 1500", searchOutput.Amount)
	}

	if _, err := Client.CancelTran(entryOutput.AccessID, entryOutput.AccessPass); err != nil {
		t.Fatal(err)
	}
	if searchOutput, err := Client.SearchTrade(orderID); err != nil {
		t.Error(err)
	} else if searchOutput.Status != "VOID" {
		t.Errorf("Status = %s; want VOID", searchOutput.Status)
	}
}

func TestPaypalOrderAPIs(t *testing.T) {
	orderID := fmt.Sprintf("%v", time.Now().UnixNano())

	entryOutput, err := Client.EntryTranPaypal(orderID, "1000", "100", "JPY")
	if err != nil {
		t.Error(err)
	}
	if _, err := Client.ExecTranPaypal(entryOutput.AccessID, entryOutput.AccessPass, orderID, "Test Order", "http://theplant.jp/gmo_redirect"); err != nil {
		t.Error(err)
	}
}
