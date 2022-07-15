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

func SaveFavourite(tool, name string, args []string) error {
	cfg, err := newToolsConfig()
	if err != nil {
		return err
	}
	return cfg.SaveFavourite(tool, name, args)
}
