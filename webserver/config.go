package main

// Copyright 2021 Matthew R. Wilson <mwilson@mattwilson.org>
//
// This file is part of virtual1403
// <https://github.com/racingmars/virtual1403>.
//
// virtual1403 is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// virtual1403 is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with virtual1403. If not, see <https://www.gnu.org/licenses/>.

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/racingmars/virtual1403/webserver/db"
	"github.com/racingmars/virtual1403/webserver/mailer"
	"github.com/racingmars/virtual1403/webserver/model"
	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	DatabaseFile            string        `yaml:"database_file"`
	CreateAdmin             string        `yaml:"create_admin"`
	FontFile                string        `yaml:"font_file"`
	ListenPort              int           `yaml:"listen_port"`
	TLSListenPort           int           `yaml:"tls_listen_port"`
	TLSDomain               string        `yaml:"tls_domain"`
	BaseURL                 string        `yaml:"server_base_url"`
	MailConfig              mailer.Config `yaml:"mail_config"`
	QuotaJobs               int           `yaml:"quota_jobs"`
	QuotaPages              int           `yaml:"quota_pages"`
	QuotaPeriod             int           `yaml:"quota_period"`
	MaxLinesPerJob          int           `yaml:"max_lines_per_job"`
	ConcurrentPrintJobs     int           `yaml:"concurrent_print_jobs"`
	InactiveMonthsCleanup   int           `yaml:"inactive_months_cleanup"`
	UnverifiedMonthsCleanup int           `yaml:"unverified_months_cleanup"`
	PDFDaysCleanup          int           `yaml:"pdf_cleanup_days"`
	NuisanceJobNames        []string      `yaml:"nuisance_job_names"`
	nuisanceJobRegex        []*regexp.Regexp
	ServerAdmin             string `yaml:"server_admin_email"`
}

func readConfig(path string) (ServerConfig, []error) {
	var c ServerConfig
	var errs []error

	f, err := os.Open(path)
	if err != nil {
		return c, []error{err}
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&c); err != nil {
		return c, []error{err}
	}

	if c.DatabaseFile == "" {
		errs = append(errs, fmt.Errorf("database file is required"))
	}

	if c.ListenPort < 1 || c.ListenPort > 65535 {
		errs = append(errs, fmt.Errorf("port number %d is invalid",
			c.ListenPort))
	}

	// TLSListenPort is optional; <= 0 we don't run TLS listener
	if c.TLSListenPort >= 1 && c.ListenPort > 65535 {
		errs = append(errs, fmt.Errorf("TLS listen port number %d is invalid",
			c.ListenPort))
	}

	// If TLSListenPort is set, we require server hostname
	if c.TLSListenPort > 0 && c.TLSDomain == "" {
		errs = append(errs, fmt.Errorf("TLS domain name is required"))
	}

	if c.BaseURL == "" {
		errs = append(errs, fmt.Errorf("server_base_url is required"))
	}

	if !mailer.ValidateAddress(c.MailConfig.FromAddress) {
		errs = append(errs,
			fmt.Errorf("address `%s` does not appear to be valid",
				c.MailConfig.FromAddress))
	}

	if !mailer.ValidateAddress(c.ServerAdmin) {
		errs = append(errs,
			fmt.Errorf("server_admin `%s` does not appear to be a "+
				"valid email address", c.ServerAdmin))
	}

	if c.MailConfig.Server == "" {
		errs = append(errs, fmt.Errorf("mail_config.server is required"))
	}
	if c.MailConfig.Port == 0 {
		errs = append(errs, fmt.Errorf("mail_config.port is required"))
	}
	if c.MailConfig.Port < 1 || c.MailConfig.Port > 65535 {
		errs = append(errs, fmt.Errorf("mail_config.port (%d) is invalid",
			c.MailConfig.Port))
	}

	if c.InactiveMonthsCleanup > 0 && c.UnverifiedMonthsCleanup <= 0 {
		errs = append(errs, fmt.Errorf("when inactive_months_cleanup is "+
			"> 0, unverified_months_cleanup must also be > 0"))
	}

	if c.UnverifiedMonthsCleanup > 0 && c.InactiveMonthsCleanup <= 0 {
		errs = append(errs, fmt.Errorf("when unverified_months_cleanup is "+
			"> 0, inactive_months_cleanup must also be > 0"))
	}

	if c.PDFDaysCleanup < 1 {
		errs = append(errs, fmt.Errorf(
			"pdf_cleanup_days is required and must be >0"))
	}

	// Parse the nuisance regular expressions
	for i := range c.NuisanceJobNames {
		r, err := regexp.Compile(c.NuisanceJobNames[i])
		if err != nil {
			errs = append(errs, fmt.Errorf("nuisance regex `%s` error: %v",
				c.NuisanceJobNames[i], err))
			continue
		}
		c.nuisanceJobRegex = append(c.nuisanceJobRegex, r)
	}

	return c, errs
}

func (a *application) createAdmin(email string) error {
	// Only proceed if admin user doesn't already exist
	_, err := a.db.GetUser(email)
	if err != db.ErrNotFound {
		log.Printf("INFO:  admin account %s already exists", email)
		return nil
	}

	// Generate random password. 128 bits; if it's good enough for AES, it's
	// good enough for us!
	pwbytes := make([]byte, 128/8)
	if n, err := rand.Read(pwbytes); err != nil || n != len(pwbytes) {
		// shouldn't be possible to have an error reading rand
		panic(err)
	}
	pwstring := hex.EncodeToString(pwbytes)

	u := model.NewUser(email, pwstring)
	u.FullName = "Administrator"
	u.Room = "ADM"
	u.Admin = true
	u.Verified = true
	u.Enabled = true

	err = a.db.SaveUser(u)
	if err != nil {
		return err
	}

	log.Printf("INFO:  Created new admin account: %s ; %s ; %s", email,
		pwstring, u.AccessKey)
	return nil
}
