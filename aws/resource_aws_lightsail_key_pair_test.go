package aws

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lightsail"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func TestAccAWSLightsailKeyPair_basic(t *testing.T) {
	var conf lightsail.KeyPair
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_lightsail_key_pair.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSLightsail(t) },
		ErrorCheck:   acctest.ErrorCheck(t, lightsail.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSLightsailKeyPairDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailKeyPairConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailKeyPairExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "fingerprint"),
					resource.TestCheckResourceAttrSet(resourceName, "public_key"),
					resource.TestCheckResourceAttrSet(resourceName, "private_key"),
				),
			},
		},
	})
}

func TestAccAWSLightsailKeyPair_publicKey(t *testing.T) {
	var conf lightsail.KeyPair
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_lightsail_key_pair.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSLightsail(t) },
		ErrorCheck:   acctest.ErrorCheck(t, lightsail.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSLightsailKeyPairDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailKeyPairConfig_imported(rName, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailKeyPairExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "fingerprint"),
					resource.TestCheckResourceAttrSet(resourceName, "public_key"),
					resource.TestCheckNoResourceAttr(resourceName, "encrypted_fingerprint"),
					resource.TestCheckNoResourceAttr(resourceName, "encrypted_private_key"),
					resource.TestCheckNoResourceAttr(resourceName, "private_key"),
				),
			},
		},
	})
}

func TestAccAWSLightsailKeyPair_encrypted(t *testing.T) {
	var conf lightsail.KeyPair
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_lightsail_key_pair.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSLightsail(t) },
		ErrorCheck:   acctest.ErrorCheck(t, lightsail.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSLightsailKeyPairDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailKeyPairConfig_encrypted(rName, testLightsailKeyPairPubKey1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailKeyPairExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "fingerprint"),
					resource.TestCheckResourceAttrSet(resourceName, "encrypted_fingerprint"),
					resource.TestCheckResourceAttrSet(resourceName, "encrypted_private_key"),
					resource.TestCheckResourceAttrSet(resourceName, "public_key"),
					resource.TestCheckNoResourceAttr(resourceName, "private_key"),
				),
			},
		},
	})
}

func TestAccAWSLightsailKeyPair_nameprefix(t *testing.T) {
	var conf1, conf2 lightsail.KeyPair

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSLightsail(t) },
		ErrorCheck:   acctest.ErrorCheck(t, lightsail.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSLightsailKeyPairDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailKeyPairConfig_prefixed(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailKeyPairExists("aws_lightsail_key_pair.lightsail_key_pair_test_omit", &conf1),
					testAccCheckAWSLightsailKeyPairExists("aws_lightsail_key_pair.lightsail_key_pair_test_prefixed", &conf2),
					resource.TestCheckResourceAttrSet("aws_lightsail_key_pair.lightsail_key_pair_test_omit", "name"),
					resource.TestCheckResourceAttrSet("aws_lightsail_key_pair.lightsail_key_pair_test_prefixed", "name"),
				),
			},
		},
	})
}

func testAccCheckAWSLightsailKeyPairExists(n string, res *lightsail.KeyPair) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No LightsailKeyPair set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn

		respKeyPair, err := conn.GetKeyPair(&lightsail.GetKeyPairInput{
			KeyPairName: aws.String(rs.Primary.Attributes["name"]),
		})

		if err != nil {
			return err
		}

		if respKeyPair == nil || respKeyPair.KeyPair == nil {
			return fmt.Errorf("KeyPair (%s) not found", rs.Primary.Attributes["name"])
		}
		*res = *respKeyPair.KeyPair
		return nil
	}
}

func testAccCheckAWSLightsailKeyPairDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lightsail_key_pair" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn

		respKeyPair, err := conn.GetKeyPair(&lightsail.GetKeyPairInput{
			KeyPairName: aws.String(rs.Primary.Attributes["name"]),
		})

		if err == nil {
			if respKeyPair.KeyPair != nil {
				return fmt.Errorf("LightsailKeyPair %q still exists", rs.Primary.ID)
			}
		}

		// Verify the error
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NotFoundException" {
				return nil
			}
		}
		return err
	}

	return nil
}

func testAccAWSLightsailKeyPairConfig_basic(lightsailName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_key_pair" "test" {
  name = "%s"
}
`, lightsailName)
}

func testAccAWSLightsailKeyPairConfig_imported(lightsailName, publicKey string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_key_pair" "test" {
  name = %[1]q

  public_key = "%[2]s"
}
`, lightsailName, publicKey)
}

func testAccAWSLightsailKeyPairConfig_encrypted(lightsailName, key string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_key_pair" "test" {
  name = %[1]q

  pgp_key = <<EOF
%[2]s
EOF
}
`, lightsailName, key)
}

func testAccAWSLightsailKeyPairConfig_prefixed() string {
	return `
resource "aws_lightsail_key_pair" "lightsail_key_pair_test_omit" {}

resource "aws_lightsail_key_pair" "lightsail_key_pair_test_prefixed" {
  name_prefix = "cts"
}
`
}

const testLightsailKeyPairPubKey1 = `mQENBFXbjPUBCADjNjCUQwfxKL+RR2GA6pv/1K+zJZ8UWIF9S0lk7cVIEfJiprzzwiMwBS5cD0da
rGin1FHvIWOZxujA7oW0O2TUuatqI3aAYDTfRYurh6iKLC+VS+F7H+/mhfFvKmgr0Y5kDCF1j0T/
063QZ84IRGucR/X43IY7kAtmxGXH0dYOCzOe5UBX1fTn3mXGe2ImCDWBH7gOViynXmb6XNvXkP0f
sF5St9jhO7mbZU9EFkv9O3t3EaURfHopsCVDOlCkFCw5ArY+DUORHRzoMX0PnkyQb5OzibkChzpg
8hQssKeVGpuskTdz5Q7PtdW71jXd4fFVzoNH8fYwRpziD2xNvi6HABEBAAG0EFZhdWx0IFRlc3Qg
S2V5IDGJATgEEwECACIFAlXbjPUCGy8GCwkIBwMCBhUIAgkKCwQWAgMBAh4BAheAAAoJEOfLr44B
HbeTo+sH/i7bapIgPnZsJ81hmxPj4W12uvunksGJiC7d4hIHsG7kmJRTJfjECi+AuTGeDwBy84TD
cRaOB6e79fj65Fg6HgSahDUtKJbGxj/lWzmaBuTzlN3CEe8cMwIPqPT2kajJVdOyrvkyuFOdPFOE
A7bdCH0MqgIdM2SdF8t40k/ATfuD2K1ZmumJ508I3gF39jgTnPzD4C8quswrMQ3bzfvKC3klXRlB
C0yoArn+0QA3cf2B9T4zJ2qnvgotVbeK/b1OJRNj6Poeo+SsWNc/A5mw7lGScnDgL3yfwCm1gQXa
QKfOt5x+7GqhWDw10q+bJpJlI10FfzAnhMF9etSqSeURBRW5AQ0EVduM9QEIAL53hJ5bZJ7oEDCn
aY+SCzt9QsAfnFTAnZJQrvkvusJzrTQ088eUQmAjvxkfRqnv981fFwGnh2+I1Ktm698UAZS9Jt8y
jak9wWUICKQO5QUt5k8cHwldQXNXVXFa+TpQWQR5yW1a9okjh5o/3d4cBt1yZPUJJyLKY43Wvptb
6EuEsScO2DnRkh5wSMDQ7dTooddJCmaq3LTjOleRFQbu9ij386Do6jzK69mJU56TfdcydkxkWF5N
ZLGnED3lq+hQNbe+8UI5tD2oP/3r5tXKgMy1R/XPvR/zbfwvx4FAKFOP01awLq4P3d/2xOkMu4Lu
9p315E87DOleYwxk+FoTqXEAEQEAAYkCPgQYAQIACQUCVduM9QIbLgEpCRDny6+OAR23k8BdIAQZ
AQIABgUCVduM9QAKCRAID0JGyHtSGmqYB/4m4rJbbWa7dBJ8VqRU7ZKnNRDR9CVhEGipBmpDGRYu
lEimOPzLUX/ZXZmTZzgemeXLBaJJlWnopVUWuAsyjQuZAfdd8nHkGRHG0/DGum0l4sKTta3OPGHN
C1z1dAcQ1RCr9bTD3PxjLBczdGqhzw71trkQRBRdtPiUchltPMIyjUHqVJ0xmg0hPqFic0fICsr0
YwKoz3h9+QEcZHvsjSZjgydKvfLYcm+4DDMCCqcHuJrbXJKUWmJcXR0y/+HQONGrGJ5xWdO+6eJi
oPn2jVMnXCm4EKc7fcLFrz/LKmJ8seXhxjM3EdFtylBGCrx3xdK0f+JDNQaC/rhUb5V2XuX6VwoH
/AtY+XsKVYRfNIupLOUcf/srsm3IXT4SXWVomOc9hjGQiJ3rraIbADsc+6bCAr4XNZS7moViAAcI
PXFv3m3WfUlnG/om78UjQqyVACRZqqAGmuPq+TSkRUCpt9h+A39LQWkojHqyob3cyLgy6z9Q557O
9uK3lQozbw2gH9zC0RqnePl+rsWIUU/ga16fH6pWc1uJiEBt8UZGypQ/E56/343epmYAe0a87sHx
8iDV+dNtDVKfPRENiLOOc19MmS+phmUyrbHqI91c0pmysYcJZCD3a502X1gpjFbPZcRtiTmGnUKd
OIu60YPNE4+h7u2CfYyFPu3AlUaGNMBlvy6PEpU=`
