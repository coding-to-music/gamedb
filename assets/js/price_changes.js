if ($('#price-changes-page').length > 0) {

    const $chosens = $('select.form-control-chosen');
    const $table = $('table.table-datatable2');
    const $form = $('form');

    // Set form fields from URL
    if (window.location.search) {
        $form.deserialize(window.location.search.substr(1));
    }

    // Setup drop downs
    $chosens.chosen({
        disable_search_threshold: 10,
        allow_single_deselect: true,
        rtl: false,
        max_selected_options: 10
    });

    // Setup Sliders
    const changeLow = $('#change-low').val();
    const changeHigh = $('#change-high').val();
    const changeElement = $('#change-slider')[0];
    const changeSlider = noUiSlider.create(changeElement, {
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

    const priceLow = $('#price-low').val();
    const priceHigh = $('#price-high').val();
    const priceElement = $('#price-slider')[0];
    const priceSlider = noUiSlider.create(priceElement, {
        start: [
            parseInt(priceLow ? priceLow : -100),
            parseInt(priceHigh ? priceHigh : 100)
        ],
        connect: true,
        step: 1,
        range: {
            'min': 0,
            'max': 100
        }
    });

    $chosens.on('change', redrawTable);
    $form.on('submit', redrawTable);
    changeSlider.on('set', onPercentChange);
    changeSlider.on('update', updateLabels);
    priceSlider.on('set', onPriceChange);
    priceSlider.on('update', updateLabels);

    function onPercentChange(e) {

        const percents = changeSlider.get();
        $('#change-low').val(percents[0]);
        $('#change-high').val(percents[1]);
        redrawTable();
    }

    function onPriceChange(e) {

        const prices = priceSlider.get();
        $('#price-low').val(prices[0]);
        $('#price-high').val(prices[1]);
        redrawTable();
    }

    function redrawTable(e) {

        // Filter out empty form fields
        let formData = $form.serializeArray();
        formData = $.grep(formData, function (v) {
            return v.value !== "";
        });

        $table.DataTable().draw();
        history.pushState({}, document.title, '/price-changes?' + $.param(formData));
        updateLabels(e);
        return false;
    }

    $(document).ready(updateLabels);

    function updateLabels(e) {

        const percents = changeSlider.get();
        const prices = priceSlider.get();

        if (percents[0] === percents[1]) {
            $('label#change-label').html('Price Change Percent (' + Math.round(percents[0]) + '%)');
        } else {
            $('label#change-label').html('Price Change Percent (' + Math.round(percents[0]) + '% - ' + Math.round(percents[1]) + '%)');
        }

        if (prices[0] === prices[1]) {
            $('label#price-label').html('Final Price (' + user.userCurrencySymbol + Math.round(prices[0]) + ')');
        } else {
            $('label#price-label').html('Final Price (' + user.userCurrencySymbol + Math.round(prices[0]) + ' - ' + user.userCurrencySymbol + Math.round(prices[1]) + ')');
        }
    }

    // Init table
    const options = $.extend(true, {}, dtDefaultOptions, {
        "order": [[4, 'desc']],
        "ajax": function (data, callback, settings) {

            data.search.type = $('#type').val();
            data.search.percents = changeSlider.get();
            data.search.prices = priceSlider.get();

            dtDefaultOptions.ajax(data, callback, settings, $(this));
        },
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
            // App/Package Name
            {
                "targets": 0,
                "render": function (data, type, row) {

                    let tagName = row[3];
                    if ($('#type').val() == 'all') {
                        if (row[0] > 0) {
                            tagName = tagName + ' <span class="badge badge-success float-right">App</span>';
                        } else if (row[1] > 0) {
                            tagName = tagName + ' <span class="badge badge-success float-right">Package</span>';
                        }
                    }

                    return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[4] + '" data-lazy-alt="' + row[3] + '"></div><div class="name">' + tagName + '</div></div>'
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
    });

    // Update table live
    const dt = $table.DataTable(options);

    websocketListener('prices', function (e) {

        const info = dt.page.info();
        if (info.page === 0) { // Page 1

            const data = $.parseJSON(e.data);
            const type = $('#type').val();

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
