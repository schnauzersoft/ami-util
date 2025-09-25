/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package aws

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sts"
)

type AMIInfo struct {
	ImageID      string
	Name         string
	CreationDate time.Time
	Owner        string
	Region       string
}

type AMIReplacement struct {
	OldAMI string
	NewAMI string
	Name   string
}

type Client struct {
	session *session.Session
	ec2     *ec2.EC2
	sts     *sts.STS
	profile string
	roleARN string
}

func NewClient(profile, roleARN string) (*Client, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile: profile,
		Config: aws.Config{
			Region: aws.String("us-east-1"), // Default region for STS
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return &Client{
		session: sess,
		ec2:     ec2.New(sess),
		sts:     sts.New(sess),
		profile: profile,
		roleARN: roleARN,
	}, nil
}

func (c *Client) AssumeRole(accountID string) (*session.Session, error) {
	roleARN := c.roleARN
	if roleARN == "" {
		roleARN = os.Getenv("AWS_ROLE_ARN")
	}
	if roleARN == "" {
		roleARN = fmt.Sprintf("arn:aws:iam::%s:role/BP-Ec2DescribeImagesRole", accountID)
	}

	sessionName := os.Getenv("AWS_ROLE_SESSION_NAME")
	if sessionName == "" {
		sessionName = "UpdateToLatestAMI"
	}

	externalID := os.Getenv("AWS_ROLE_EXTERNAL_ID")

	creds := stscreds.NewCredentials(c.session, roleARN, func(p *stscreds.AssumeRoleProvider) {
		p.RoleSessionName = sessionName
		if externalID != "" {
			p.ExternalID = aws.String(externalID)
		}
	})

	sess, err := session.NewSession(&aws.Config{
		Credentials: creds,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create session with assumed role: %w", err)
	}

	return sess, nil
}

func (c *Client) GetLatestAMIs(accountID, region string, patterns []string) ([]AMIReplacement, error) {
	sess, err := c.AssumeRole(accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to assume role for account %s: %w", accountID, err)
	}

	ec2Client := ec2.New(sess, &aws.Config{Region: aws.String(region)})
	var replacements []AMIReplacement

	for _, pattern := range patterns {
		amis, err := c.findAMIsByPattern(ec2Client, accountID, pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to find AMIs for pattern %s: %w", pattern, err)
		}

		if len(amis) == 0 {
			continue
		}

		sort.Slice(amis, func(i, j int) bool {
			return amis[i].CreationDate.After(amis[j].CreationDate)
		})

		latest := amis[0]
		for _, ami := range amis[1:] {
			replacements = append(replacements, AMIReplacement{
				OldAMI: ami.ImageID,
				NewAMI: latest.ImageID,
				Name:   ami.Name,
			})
		}
	}

	return replacements, nil
}

func (c *Client) findAMIsByPattern(ec2Client *ec2.EC2, owner, pattern string) ([]AMIInfo, error) {
	input := &ec2.DescribeImagesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("name"),
				Values: []*string{aws.String(pattern)},
			},
		},
		Owners: []*string{aws.String(owner)},
	}

	result, err := ec2Client.DescribeImages(input)
	if err != nil {
		return nil, err
	}

	var amis []AMIInfo
	for _, image := range result.Images {
		creationDate, err := time.Parse(time.RFC3339, *image.CreationDate)
		if err != nil {
			continue
		}

		amis = append(amis, AMIInfo{
			ImageID:      *image.ImageId,
			Name:         *image.Name,
			CreationDate: creationDate,
			Owner:        owner,
		})
	}

	return amis, nil
}

func ExtractAMIPatterns(content string) []string {
	amiRegex := regexp.MustCompile(`ami-[a-f0-9]{8,17}`)
	amiMatches := amiRegex.FindAllString(content, -1)

	amiMap := make(map[string]bool)
	for _, ami := range amiMatches {
		amiMap[ami] = true
	}

	var amis []string
	for ami := range amiMap {
		amis = append(amis, ami)
	}

	return amis
}

func ReplaceAMIsInContent(content string, replacements []AMIReplacement) (string, int) {
	replaceCount := 0
	newContent := content

	for _, replacement := range replacements {
		oldCount := strings.Count(newContent, replacement.OldAMI)
		if oldCount > 0 {
			newContent = strings.ReplaceAll(newContent, replacement.OldAMI, replacement.NewAMI)
			replaceCount += oldCount
		}
	}

	return newContent, replaceCount
}
