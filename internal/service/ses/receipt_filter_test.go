package ses_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccAWSSESReceiptFilter_basic(t *testing.T) {
	resourceName := "aws_ses_receipt_filter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSSES(t); testAccPreCheckSESReceiptRule(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ses.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSESReceiptFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESReceiptFilterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESReceiptFilterExists(resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "ses", fmt.Sprintf("receipt-filter/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cidr", "10.10.10.10"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy", "Block"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSESReceiptFilter_disappears(t *testing.T) {
	resourceName := "aws_ses_receipt_filter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSSES(t); testAccPreCheckSESReceiptRule(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ses.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSESReceiptFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESReceiptFilterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESReceiptFilterExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceReceiptFilter(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSESReceiptFilterDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_receipt_filter" {
			continue
		}

		response, err := conn.ListReceiptFilters(&ses.ListReceiptFiltersInput{})
		if err != nil {
			return err
		}

		for _, element := range response.Filters {
			if aws.StringValue(element.Name) == rs.Primary.ID {
				return fmt.Errorf("SES Receipt Filter (%s) still exists", rs.Primary.ID)
			}
		}
	}

	return nil

}

func testAccCheckAwsSESReceiptFilterExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES receipt filter not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES receipt filter ID not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESConn

		response, err := conn.ListReceiptFilters(&ses.ListReceiptFiltersInput{})
		if err != nil {
			return err
		}

		for _, element := range response.Filters {
			if aws.StringValue(element.Name) == rs.Primary.ID {
				return nil
			}
		}

		return fmt.Errorf("The receipt filter was not created")
	}
}

func testAccAWSSESReceiptFilterConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_filter" "test" {
  cidr   = "10.10.10.10"
  name   = %q
  policy = "Block"
}
`, rName)
}