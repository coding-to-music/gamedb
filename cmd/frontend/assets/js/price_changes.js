if ($('#price-changes-page').length > 0) {

    // Setup drop downs
    $('select.form-control-chosen').chosen({
        disable_search_threshold: 5,
        allow_single_deselect: true,
        max_selected_options: 10
    });

    // Change slider
    const changeLow = $('#change-low').val();
    const changeHigh = $('#change-high').val();
    const $changeElement = $('#change-slider');
    const changeSlider = noUiSlider.create($changeElement[0], {
        start: [
            parseInt(changeLow ? changeLow : -100),
            parseInt(changeHigh ? changeHigh : 100)
        ],
        connect: true,
        step: 1,
        range: {
            'min': -100,
            'max': 100
        }
    });

    // Price slider
    const priceLow = $('#price-low').val();
    const priceHigh = $('#price-high').val();
    const $priceElement = $('#price-slider');
    const priceSlider = noUiSlider.create($priceElement[0], {
        start: [
            parseInt(priceLow ? priceLow : 0),
            parseInt(priceHigh ? priceHigh : 100)
        ],
        connect: true,
        step: 1,
        range: {
            'min': 0,
            'max': 100
        }
    });

    //
    function updateLabels(e) {

        const percents = changeSlider.get();
        if (percents[0] === percents[1]) {
            $('label#change-label').html('Price Change Percent (' + Math.round(percents[0]) + '%)');
        } else {
            $('label#change-label').html('Price Change Percent (' + Math.round(percents[0]) + '% - ' + Math.round(percents[1]) + '%)');
        }

        const prices = priceSlider.get();

        let left = Math.round(prices[0]);
        left = left === 100 ? '100+' : left;

        let right = Math.round(prices[1]);
        right = right === 100 ? '100+' : right;

        if (prices[0] === prices[1]) {
            $('label#price-label').html('Final Price (' + user.userCurrencySymbol + left + ')');
        } else {
            $('label#price-label').html('Final Price (' + user.userCurrencySymbol + left + ' - ' + user.userCurrencySymbol + right + ')');
        }
    }

    window.updateLabels = updateLabels;

    $(updateLabels);

    $typeField = $('#type');

    // Init table
    const options = {
        "order": [[4, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-app-id', data[0]);
            $(row).attr('data-link', data[5]);

            let x;
            if (data[9]) {
                x = Math.min(data[9], 100); // Get a range of -100 to 100
                x += 100; // Get a range of 0 to 200
                x = x / 2; // Get a range of 0 to 100
            } else {
                x = 100; // Infinite price increase
            }

            $(row).addClass('col-grad-' + Math.round(x));
        },
        "columnDefs": [
            // App Name / Package Name
            {
                "targets": 0,
                "render": function (data, type, row) {

                    let tagName = row[3];
                    if ($typeField.val() === 'all') {
                        if (row[0] > 0) {
                            tagName = tagName + ' <span class="badge badge-success float-right">App</span>';
                        } else if (row[1] > 0) {
                            tagName = tagName + ' <span class="badge badge-success float-right">Package</span>';
                        }
                    }

                    return '<a href="' + row[5] + '" class="icon-name"><div class="icon"><img data-lazy="' + row[4] + '" alt="" data-lazy-alt="' + row[3] + '"></div><div class="name">' + tagName + '</div></a>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img')
                },
                "orderable": false
            },
            // Before
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[6];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false
            },
            // After
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[7];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false
            },
            // Change
            {
                "targets": 3,
                "render": function (data, type, row) {

                    const small = '<small>' + row[9] + '%</small>';

                    if (row[9] === 0) {
                        return row[8] + ' <small>∞%</small> ';
                    }

                    return row[8] + ' ' + small;
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false
            },
            // Time
            {
                "targets": 4,
                "render": function (data, type, row) {
                    return '<span data-toggle="tooltip" data-placement="left" title="' + row[10] + '" data-livestamp="' + row[11] + '">' + row[10] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false
            }
        ]
    };

    // Update table live
    const searchFields = [
        $typeField,
        $changeElement,
        $priceElement,
    ];

    const $table = $('table.table');
    const dt = $table.gdbTable({tableOptions: options, searchFields: searchFields});

    websocketListener('prices', function (e) {

        const info = dt.page.info();
        if (info.page === 0) { // Page 1

            const data = JSON.parse(e.data);
            const type = $typeField.val();

            // Check cc matches
            if (data.Data[13] === user.prodCC) {
                // Check product type
                if (type === 'all' || (type === 'apps' && data.Data[0] > 0) || (type === 'packages' && data.Data[1] > 0)) {
                    // Add row
                    addDataTablesRow(options, data.Data, info.length, $table);
                }
            }
        }
    });
}
