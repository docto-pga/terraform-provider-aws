package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func init() {
	resource.AddTestSweepers("aws_route53_query_log", &resource.Sweeper{
		Name: "aws_route53_query_log",
		F:    testSweepRoute53QueryLogs,
	})
}

func testSweepRoute53QueryLogs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).Route53Conn
	var sweeperErrs *multierror.Error

	err = conn.ListQueryLoggingConfigsPages(&route53.ListQueryLoggingConfigsInput{}, func(page *route53.ListQueryLoggingConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, queryLoggingConfig := range page.QueryLoggingConfigs {
			id := aws.StringValue(queryLoggingConfig.Id)

			r := ResourceQueryLog()
			d := r.Data(nil)
			d.SetId(id)
			err := r.Delete(d, client)
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Route53 query logging configuration (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})
	// In unsupported AWS partitions, the API may return an error even the SDK cannot handle.
	// Reference: https://github.com/aws/aws-sdk-go/issues/3313
	if sweep.SkipSweepError(err) || tfawserr.ErrMessageContains(err, "SerializationError", "failed to unmarshal error message") || tfawserr.ErrMessageContains(err, "AccessDeniedException", "Unable to determine service/operation name to be authorized") {
		log.Printf("[WARN] Skipping Route53 query logging configurations sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Route53 query logging configurations: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSRoute53QueryLog_basic(t *testing.T) {
	cloudwatchLogGroupResourceName := "aws_cloudwatch_log_group.test"
	resourceName := "aws_route53_query_log.test"
	route53ZoneResourceName := "aws_route53_zone.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	var queryLoggingConfig route53.QueryLoggingConfig
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckRoute53QueryLog(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRoute53QueryLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSRoute53QueryLogResourceConfigBasic1(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53QueryLogExists(resourceName, &queryLoggingConfig),
					acctest.MatchResourceAttrGlobalARNNoAccount(resourceName, "arn", "route53", regexp.MustCompile("queryloggingconfig/.+")),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_log_group_arn", cloudwatchLogGroupResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "zone_id", route53ZoneResourceName, "zone_id"),
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

func TestAccAWSRoute53QueryLog_disappears(t *testing.T) {
	resourceName := "aws_route53_query_log.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	var queryLoggingConfig route53.QueryLoggingConfig
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckRoute53QueryLog(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRoute53QueryLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSRoute53QueryLogResourceConfigBasic1(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53QueryLogExists(resourceName, &queryLoggingConfig),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceQueryLog(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRoute53QueryLog_disappears_hostedZone(t *testing.T) {
	resourceName := "aws_route53_query_log.test"
	route53ZoneResourceName := "aws_route53_zone.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	var queryLoggingConfig route53.QueryLoggingConfig
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckRoute53QueryLog(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRoute53QueryLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSRoute53QueryLogResourceConfigBasic1(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53QueryLogExists(resourceName, &queryLoggingConfig),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceZone(), route53ZoneResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRoute53QueryLogExists(pr string, queryLoggingConfig *route53.QueryLoggingConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProviderRoute53QueryLog.Meta().(*conns.AWSClient).Route53Conn
		rs, ok := s.RootModule().Resources[pr]
		if !ok {
			return fmt.Errorf("Not found: %s", pr)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		out, err := conn.GetQueryLoggingConfig(&route53.GetQueryLoggingConfigInput{
			Id: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}
		if out.QueryLoggingConfig == nil {
			return fmt.Errorf("Route53 query logging configuration does not exist: %q", rs.Primary.ID)
		}

		*queryLoggingConfig = *out.QueryLoggingConfig

		return nil
	}
}

func testAccCheckRoute53QueryLogDestroy(s *terraform.State) error {
	conn := testAccProviderRoute53QueryLog.Meta().(*conns.AWSClient).Route53Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_query_log" {
			continue
		}

		out, err := conn.GetQueryLoggingConfig(&route53.GetQueryLoggingConfigInput{
			Id: aws.String(rs.Primary.ID),
		})

		if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchQueryLoggingConfig) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading Route 53 Query Logging Configuration (%s): %w", rs.Primary.ID, err)
		}

		if out.QueryLoggingConfig != nil {
			return fmt.Errorf("Route53 query logging configuration exists: %q", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSRoute53QueryLogResourceConfigBasic1(rName, domainName string) string {
	return acctest.ConfigCompose(
		testAccRoute53QueryLogRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name              = "/aws/route53/${aws_route53_zone.test.name}"
  retention_in_days = 1
}

data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["arn:${data.aws_partition.current.partition}:logs:*:*:log-group:/aws/route53/*"]

    principals {
      identifiers = ["route53.${data.aws_partition.current.dns_suffix}"]
      type        = "Service"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_name     = %[1]q
  policy_document = data.aws_iam_policy_document.test.json
}

resource "aws_route53_zone" "test" {
  name = %[2]q
}

resource "aws_route53_query_log" "test" {
  depends_on = [aws_cloudwatch_log_resource_policy.test]

  cloudwatch_log_group_arn = aws_cloudwatch_log_group.test.arn
  zone_id                  = aws_route53_zone.test.zone_id
}
`, rName, domainName))
}
