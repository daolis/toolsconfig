/*
 * [y] hybris Platform
 *
 * Copyright (c) 2022 SAP SE or an SAP affiliate company. All rights reserved.
 *
 * This software is the confidential and proprietary information of SAP
 * ("Confidential Information"). You shall not disclose such Confidential
 * Information and shall use it only in accordance with the terms of the
 * license agreement you entered into with SAP.
 */

package commands

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/ttacon/chalk"
)

func ListFavourites(tool string) error {
	cfg, err := newToolsConfig()
	if err != nil {
		return err
	}
	fmt.Printf("\nFavourites (execute with '%s --run [NAME])\n", tool)
	w := tabwriter.NewWriter(os.Stdout, 0, 2, 4, ' ', 0)
	_, _ = fmt.Fprintf(w, "NAME\tCOMMAND\n")
	for _, fav := range cfg.GetFavourites(tool) {
		_, _ = fmt.Fprintf(w, "%s%s%s\t'%s %s'\n", chalk.Yellow, fav.Name, chalk.ResetColor, tool, strings.Join(fav.Args, " "))
	}
	_ = w.Flush()
	return nil
}
