if ($('#bundles-page').length > 0) {

    // Setup drop downs
    $('select.form-control-chosen').chosen({
        disable_search_threshold: 5,
        allow_single_deselect: true,
        max_selected_options: 10
    });

    // Discount
    const $discountElement = $('#discount');
    const discountSlider = noUiSlider.create($discountElement[0], {
        start: [0, 100],
        connect: true,
        step: 1,
        range: {
            'min': 0,
            'max': 100,
        },
        format: {
            to: (v) => parseFloat(v).toFixed(0),
            from: (v) => parseFloat(v).toFixed(0)
        },
    });

    // Apps slider
    const $appsElement = $('#apps');
    const appsSlider = noUiSlider.create($appsElement[0], {
        start: [0, 100],
        connect: true,
        step: 1,
        range: {
            'min': 0,
            'max': 100,
        },
        format: {
            to: (v) => parseFloat(v).toFixed(0),
            from: (v) => parseFloat(v).toFixed(0)
        },
    });

    // Packages slider
    const $packagesElement = $('#packages');
    const packagesSlider = noUiSlider.create($packagesElement[0], {
        start: [0, 100],
        connect: true,
        step: 1,
        range: {
            'min': 0,
            'max': 100,
        },
        format: {
            to: (v) => parseFloat(v).toFixed(0),
            from: (v) => parseFloat(v).toFixed(0)
        },
    });

    //
    function updateLabels(e) {

        const discount = discountSlider.get();
        if (discount[0] === discount[1]) {
            $('label[for=discount]').html('Discount (' + Math.round(discount[0]) + ')');
        } else {
            $('label[for=discount]').html('Discount (' + Math.round(discount[0]) + ' - ' + Math.round(discount[1]) + ')');
        }

        //
        const apps = appsSlider.get();
        const appsRight = (apps[1] === '100' ? '100+' : apps[1]);
        if (apps[0] === apps[1]) {
            $('label[for=apps]').html('Apps (' + appsRight + ')');
        } else {
            $('label[for=apps]').html('Apps (' + apps[0] + ' - ' + appsRight + ')');
        }

        //
        const packages = packagesSlider.get();
        const packagesRight = (apps[1] === '100' ? '100+' : apps[1]);
        if (packages[0] === packages[1]) {
            $('label[for=packages]').html('Packages (' + packagesRight + ')');
        } else {
            $('label[for=packages]').html('Packages (' + apps[0] + ' - ' + packagesRight + ')');
        }
    }

    window.updateLabels = updateLabels;

    $(updateLabels);

    //
    const options = {
        "order": [[5, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-link', data[2]);
        },
        "columnDefs": [
            // Icon / Bundle Name
            {
                "targets": 0,
                "render": function (data, type, row) {

                    let tagName = row[1];
                    if (row[7]) {
                        tagName = tagName + ' <span class="badge badge-success">Lowest</span>';
                    }

                    return '<a href="' + row[2] + '" class="icon-name"><div class="icon"><img src="/assets/img/no-app-image-square.jpg" alt="' + row[1] + '"></div><div class="name">' + tagName + '</div></a>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                },
                "orderable": false,
            },
            // Discount
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[4] + '%'
                },
                "orderSequence": ["asc", "desc"],
            },
            // Price
            {
                "targets": 2,
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "render": function (data, type, row) {
                    if (user.prodCC in row[9]) {
                        return row[9][user.prodCC];
                    }
                    return '-';
                },
                "orderable": false,
            },
            // Apps
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[5].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // Packages
            {
                "targets": 4,
                "render": function (data, type, row) {
                    return row[6].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // Updated At
            {
                "targets": 5,
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "render": function (data, type, row) {
                    return '<span data-livestamp="' + row[3] + '"></span>';
                }
            },
            // Link
            {
                "targets": 6,
                "render": function (data, type, row) {
                    if (row[8]) {
                        return '<a href="' + row[8] + '" target="_blank" rel="noopener"><i class="fas fa-link"></i></a>';
                    }
                    return '';
                },
                "orderable": false,
            },
            // Search score
            {
                "targets": 7,
                "render": function (data, type, row) {
                    return row[10];
                },
                "orderable": false,
                "visible": user.isLocal,
            },
        ]
    };


    const $table = $('table.table');
    const searchFields = [
        $('#search'),
        $discountElement,
        $appsElement,
        $packagesElement,
        $('#type'),
        $('#giftable'),
        $('#onsale'),
    ];

    const dt = $table.gdbTable({tableOptions: options, searchFields: searchFields});

    websocketListener('bundles', function (e) {

        const info = dt.page.info();
        if (info.page === 0) { // Page 1

            const data = JSON.parse(e.data);
            addDataTablesRow(options, data.Data, info.length, $table);
        }
    });
}
