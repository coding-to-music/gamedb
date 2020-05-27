if ($('#groups-page').length > 0) {

    // Setup drop downs
    $('select.form-control-chosen').chosen({
        disable_search_threshold: 10,
        allow_single_deselect: true,
        rtl: false,
        max_selected_options: 1
    });

    //
    const options = {
        "order": [[2, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-group-id', data[0]);
            $(row).attr('data-link', data[10]);
        },
        "columnDefs": [
            // Rank
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return row[9].toLocaleString();
                },
                "orderable": false,
            },
            // Icon / Name
            {
                "targets": 1,
                "render": function (data, type, row) {

                    let name = row[1] + '<br><small>' + row[4] + '</small>';

                    if (row[8]) {
                        name += '<span class="badge badge-danger float-right">Removed</span>';
                    }

                    return '<div class="icon-name"><div class="icon"><img class="tall" data-src="/assets/img/no-app-image-square.jpg" data-lazy="' + row[5] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name markable">' + name + '</div></div>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                },
                "orderable": false,
            },
            // Members
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[6].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // Trend Value
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[7];
                },
                "orderSequence": ["desc", "asc"],
            },
            // Search Score
            {
                "targets": 4,
                "render": function (data, type, row) {
                    return row[11].toLocaleString();
                },
                "orderable": false,
                "visible": false,
            },
            // Link
            {
                "targets": 5,
                "render": function (data, type, row) {
                    if (row[2]) {
                        return '<a href="' + row[2] + '" target="_blank" rel="noopener"><i class="fas fa-link"></i></a>';
                    }
                    return '';
                },
                "orderable": false,
            },
        ]
    };

    $('table.table').gdbTable({
        tableOptions: options,
        searchFields: [
            $('#search'),
            $('#filter'),
        ],
    });
}
