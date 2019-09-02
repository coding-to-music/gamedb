package tasks

type DevCodeRun struct {
}

func (c DevCodeRun) ID() string {
	return "run-dev-code"
}

func (c DevCodeRun) Name() string {
	return "Run dev code"
}

func (c DevCodeRun) Cron() string {
	return ""
}

func (c DevCodeRun) work() {

	// var err error
	// gorm, err := sql.GetMySQLClient()
	// if err != nil {
	// 	log.Err(err)
	// 	return
	// }
	//
	// gorm = gorm.Select([]string{"*"}) // Need everything so when we save we dont lose data
	// gorm = gorm.Order("id asc")
	//
	// // Get rows
	// var bundles []sql.Bundle
	// gorm = gorm.Find(&bundles)
	// if gorm.Error != nil {
	// 	log.Err(gorm.Error)
	// 	return
	// }
	//
	// for _, bundle := range bundles {
	//
	// 	builder := influxql.NewBuilder()
	// 	builder.AddSelect("discount", "")
	// 	builder.SetFrom(helpers.InfluxGameDB, helpers.InfluxRetentionPolicyAllTime.String(), helpers.InfluxMeasurementApps.String())
	// 	builder.AddWhere("bundle_id", "=", bundle.ID)
	//
	// 	resp, err := helpers.InfluxQuery(builder.String())
	// 	if err != nil {
	// 		log.Err(err)
	// 		continue
	// 	}
	//
	// 	if len(resp.Results) == 0 || len(resp.Results[0].Series) == 0 || len(resp.Results[0].Series[0].Values) == 0 {
	// 		log.Info(bundle.ID, "skipped")
	// 		continue
	// 	}
	//
	// 	type price struct {
	// 		Time    time.Time `json:"time"`
	// 		Percent int       `json:"percent"`
	// 	}
	//
	// 	var prices []price
	// 	var priceDocuments []mongo.Document
	//
	// 	// Convert influx response to slice
	// 	for _, influxRow := range resp.Results[0].Series[0].Values {
	//
	// 		t, err := time.Parse(time.RFC3339, influxRow[0].(string))
	// 		if err != nil {
	// 			log.Err(err)
	// 			continue
	// 		}
	//
	// 		i, err := strconv.Atoi(influxRow[1].(json.Number).String())
	// 		if err != nil {
	// 			log.Err(err)
	// 			continue
	// 		}
	//
	// 		prices = append(prices, price{Time: t, Percent: i})
	// 	}
	//
	// 	if len(prices) == 0 {
	// 		log.Info(bundle.ID, "no prices")
	// 		continue
	// 	}
	//
	// 	// Sort prices, oldest first
	// 	sort.Slice(prices, func(i, j int) bool {
	// 		return prices[i].Time.Unix() < prices[j].Time.Unix()
	// 	})
	//
	// 	// Save to mongo
	// 	var last = 1 // A value that will never match the first price
	// 	for _, v := range prices {
	//
	// 		bundle.SetDiscount(v.Percent)
	//
	// 		if v.Percent != last {
	//
	// 			document := mongo.BundlePrice{
	// 				CreatedAt: v.Time,
	// 				BundleID:  bundle.ID,
	// 				Discount:  v.Percent,
	// 			}
	//
	// 			priceDocuments = append(priceDocuments, document)
	//
	// 			last = v.Percent
	// 		}
	// 	}
	//
	// 	_, err = mongo.InsertDocuments(mongo.CollectionBundlePrices, priceDocuments)
	//
	// 	err = bundle.Save()
	// 	if err != nil {
	// 		log.Err(err)
	// 		continue
	// 	}
	// }
	//
}
