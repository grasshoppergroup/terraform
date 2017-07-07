package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsAutoscalingAttachment_elb(t *testing.T) {

	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoscalingAttachment_elb(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutocalingElbAttachmentExists("aws_autoscaling_group.asg", 0),
				),
			},
			{
				Config: testAccAWSAutoscalingAttachment_elb_associated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutocalingElbAttachmentExists("aws_autoscaling_group.asg", 1),
				),
			},
			{
				Config: testAccAWSAutoscalingAttachment_elb_double_associated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutocalingElbAttachmentExists("aws_autoscaling_group.asg", 2),
				),
			},
			{
				Config: testAccAWSAutoscalingAttachment_elb_associated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutocalingElbAttachmentExists("aws_autoscaling_group.asg", 1),
				),
			},
			{
				Config: testAccAWSAutoscalingAttachment_elb(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutocalingElbAttachmentExists("aws_autoscaling_group.asg", 0),
				),
			},
		},
	})
}

func TestAccAwsAutoscalingAttachment_albTargetGroup(t *testing.T) {

	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoscalingAttachment_alb(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutocalingAlbAttachmentExists("aws_autoscaling_group.asg", 0),
				),
			},
			{
				Config: testAccAWSAutoscalingAttachment_alb_associated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutocalingAlbAttachmentExists("aws_autoscaling_group.asg", 1),
				),
			},
			{
				Config: testAccAWSAutoscalingAttachment_alb_double_associated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutocalingAlbAttachmentExists("aws_autoscaling_group.asg", 2),
				),
			},
			{
				Config: testAccAWSAutoscalingAttachment_alb_associated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutocalingAlbAttachmentExists("aws_autoscaling_group.asg", 1),
				),
			},
			{
				Config: testAccAWSAutoscalingAttachment_alb(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutocalingAlbAttachmentExists("aws_autoscaling_group.asg", 0),
				),
			},
		},
	})
}

func testAccCheckAWSAutocalingElbAttachmentExists(asgname string, loadBalancerCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[asgname]
		if !ok {
			return fmt.Errorf("Not found: %s", asgname)
		}

		conn := testAccProvider.Meta().(*AWSClient).autoscalingconn
		asg := rs.Primary.ID

		actual, err := conn.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{aws.String(asg)},
		})

		if err != nil {
			return fmt.Errorf("Received an error when attempting to load %s:  %s", asg, err)
		}

		if loadBalancerCount != len(actual.AutoScalingGroups[0].LoadBalancerNames) {
			return fmt.Errorf("Error: ASG has the wrong number of load balacners associated.  Expected [%d] but got [%d]", loadBalancerCount, len(actual.AutoScalingGroups[0].LoadBalancerNames))
		}

		return nil
	}
}

func testAccCheckAWSAutocalingAlbAttachmentExists(asgname string, targetGroupCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[asgname]
		if !ok {
			return fmt.Errorf("Not found: %s", asgname)
		}

		conn := testAccProvider.Meta().(*AWSClient).autoscalingconn
		asg := rs.Primary.ID

		actual, err := conn.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{aws.String(asg)},
		})

		if err != nil {
			return fmt.Errorf("Recieved an error when attempting to load %s:  %s", asg, err)
		}

		if targetGroupCount != len(actual.AutoScalingGroups[0].TargetGroupARNs) {
			return fmt.Errorf("Error: ASG has the wrong number of Target Groups associated.  Expected [%d] but got [%d]", targetGroupCount, len(actual.AutoScalingGroups[0].TargetGroupARNs))
		}

		return nil
	}
}

func testAccAWSAutoscalingAttachment_alb(rInt int) string {
	return fmt.Sprintf(`
resource "aws_alb_target_group" "test" {
  name = "test-alb-%d"
  port = 443
  protocol = "HTTPS"
  vpc_id = "${aws_vpc.test.id}"

  deregistration_delay = 200

  stickiness {
    type = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path = "/health"
    interval = 60
    port = 8081
    protocol = "HTTP"
    timeout = 3
    healthy_threshold = 3
    unhealthy_threshold = 3
    matcher = "200-299"
  }

  tags {
    TestName = "TestAccAWSALBTargetGroup_basic"
  }
}

resource "aws_alb_target_group" "another_test" {
  name = "atest-alb-%d"
  port = 443
  protocol = "HTTPS"
  vpc_id = "${aws_vpc.test.id}"

  deregistration_delay = 200

  stickiness {
    type = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path = "/health"
    interval = 60
    port = 8081
    protocol = "HTTP"
    timeout = 3
    healthy_threshold = 3
    unhealthy_threshold = 3
    matcher = "200-299"
  }

  tags {
    TestName = "TestAccAWSALBTargetGroup_basic"
  }
}

resource "aws_autoscaling_group" "asg" {
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]
  name = "asg-lb-assoc-terraform-test_%d"
  max_size = 1
  min_size = 0
  desired_capacity = 0
  health_check_grace_period = 300
  force_delete = true
  launch_configuration = "${aws_launch_configuration.as_conf.name}"

  tag {
    key = "Name"
    value = "terraform-asg-lg-assoc-test"
    propagate_at_launch = true
  }
}

resource "aws_launch_configuration" "as_conf" {
    name = "test_config_%d"
    image_id = "ami-f34032c3"
    instance_type = "t1.micro"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags {
    TestName = "TestAccAWSALBTargetGroup_basic"
  }
}
`, rInt, rInt, rInt, rInt)
}

func testAccAWSAutoscalingAttachment_elb(rInt int) string {
	return fmt.Sprintf(`
resource "aws_security_group" "tf_open_ingress" {
  name        = "tf_open_ingress_sg"
  description = "tf_open_ingress_sg"
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  tags {
    Name = "testAccAWSAutoscalingAttachment_elb"
  }
}

resource "aws_elb" "foo" {
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  health_check {
    healthy_threshold = 2
    unhealthy_threshold = 2
    target = "HTTP:80/"
    interval = 5
    timeout = 2
  }
}

resource "aws_elb" "bar" {
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]

  listener {
    # NOTE: This is an INTENTIONALLY misconfigured port to ensure that the waitForAsg
    # process will only take into account the LBs/attachments that are explicitly
    # called out in the attachment resource
    instance_port     = 8000
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  health_check {
    healthy_threshold = 2
    unhealthy_threshold = 2
    target = "HTTP:80/"
    interval = 5
    timeout = 2
  }
}

resource "aws_launch_configuration" "as_conf" {
  // need an AMI that listens on :80 at boot, this is:
  // bitnami-nginxstack-1.6.1-0-linux-ubuntu-14.04.1-x86_64-hvm-ebs-ami-99f5b1a9-3
  image_id = "ami-b5b3fc85"
  instance_type = "t2.micro"
  security_groups = ["${aws_security_group.tf_open_ingress.id}"]
}

resource "aws_autoscaling_group" "asg" {
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]
  name = "asg-lb-assoc-terraform-test_%d"
  max_size = 2
  min_size = 0
  desired_capacity = 2
  health_check_grace_period = 300
  force_delete = true
  launch_configuration = "${aws_launch_configuration.as_conf.name}"

  tag {
    key = "Name"
    value = "terraform-asg-lg-assoc-test"
    propagate_at_launch = true
  }
}`, rInt, rInt)
}

func testAccAWSAutoscalingAttachment_elb_associated(rInt int) string {
	return testAccAWSAutoscalingAttachment_elb(rInt) + `
resource "aws_autoscaling_attachment" "asg_attachment_foo" {
  autoscaling_group_name = "${aws_autoscaling_group.asg.id}"
  elb                    = "${aws_elb.foo.id}"
  wait_for_elb_capacity  = 2
}`
}

func testAccAWSAutoscalingAttachment_alb_associated(rInt int) string {
	return testAccAWSAutoscalingAttachment_alb(rInt) + `
resource "aws_autoscaling_attachment" "asg_attachment_foo" {
  autoscaling_group_name = "${aws_autoscaling_group.asg.id}"
  alb_target_group_arn   = "${aws_alb_target_group.test.arn}"
}`
}

func testAccAWSAutoscalingAttachment_elb_double_associated(rInt int) string {
	return testAccAWSAutoscalingAttachment_elb_associated(rInt) + `
resource "aws_autoscaling_attachment" "asg_attachment_bar" {
  autoscaling_group_name = "${aws_autoscaling_group.asg.id}"
  elb                    = "${aws_elb.bar.id}"
  wait_for_elb_capacity  = 2
}`
}

func testAccAWSAutoscalingAttachment_alb_double_associated(rInt int) string {
	return testAccAWSAutoscalingAttachment_alb_associated(rInt) + `
resource "aws_autoscaling_attachment" "asg_attachment_bar" {
  autoscaling_group_name = "${aws_autoscaling_group.asg.id}"
  alb_target_group_arn   = "${aws_alb_target_group.another_test.arn}"
}`
}
