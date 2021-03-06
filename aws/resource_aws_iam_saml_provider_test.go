package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSIAMSamlProvider_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_saml_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIAMSamlProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMSamlProviderConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMSamlProviderExists(resourceName),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "iam", regexp.MustCompile(`saml-provider/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "saml_metadata_document"),
				),
			},
			{
				Config: testAccIAMSamlProviderConfigUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMSamlProviderExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "saml_metadata_document"),
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

func TestAccAWSIAMSamlProvider_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_saml_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIAMSamlProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMSamlProviderConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMSamlProviderExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsIamSamlProvider(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckIAMSamlProviderDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_saml_provider" {
			continue
		}

		input := &iam.GetSAMLProviderInput{
			SAMLProviderArn: aws.String(rs.Primary.ID),
		}
		out, err := conn.GetSAMLProvider(input)

		if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if out != nil {
			return fmt.Errorf("IAM SAML Provider (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckIAMSamlProviderExists(id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("Not Found: %s", id)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).iamconn
		_, err := conn.GetSAMLProvider(&iam.GetSAMLProviderInput{
			SAMLProviderArn: aws.String(rs.Primary.ID),
		})

		return err
	}
}

func testAccIAMSamlProviderConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %q
  saml_metadata_document = "${file("./test-fixtures/saml-metadata.xml")}"
}
`, rName)
}

func testAccIAMSamlProviderConfigUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %q
  saml_metadata_document = "${file("./test-fixtures/saml-metadata-modified.xml")}"
}
`, rName)
}
