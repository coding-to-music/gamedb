if ($('#packages-page').length > 0) {

    // Setup drop downs
    $('select.form-control-chosen').chosen({
        disable_search_threshold: 10,
        allow_single_deselect: true,
        rtl: false,
        max_selected_options: 10
    });

    const options = {
        "order": [[4, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-link', data[1]);
        },
        "columnDefs": [
            // Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[8] + '" data-lazy-alt="' + row[2] + '"></div><div class="name">' + row[2] + '</div></div>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                },
                "orderable": false,
            },
            // Price
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[5];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                'orderSequence': ['desc', 'asc'],
            },
            // Discount
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[9];
                },
                'orderSequence': ['desc', 'asc'],
            },
            // Apps
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[4].toLocaleString();
                },
                'orderSequence': ['desc', 'asc'],
            },
            // Updated Time
            {
                "targets": 4,
                "render": function (data, type, row) {
                    return '<span data-toggle="tooltip" data-placement="left" title="' + row[7] + '" data-livestamp="' + row[6] + '">' + row[7] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                'orderSequence': ['desc'],
            },
            // Link
            {
                "targets": 5,
                "render": function (data, type, row) {
                    if (row[10]) {
                        return '<a href="' + row[10] + '" target="_blank" rel="nofollow"><i class="fas fa-link"></i></a>';
                    }
                    return '';
                },
                "orderable": false,
            },
        ]
    };

    const searchFields = [
        $('#status'),
        $('#platform'),
        $('#license'),
        $('#billing'),
    ];

    const $table = $('table.table');
    const dt = $table.gdbTable({
        tableOptions: options,
        searchFields: searchFields
    });

    websocketListener('packages', function (e) {

        const info = dt.page.info();
        if (info.page === 0) { // Page 1

            const data = JSON.parse(e.data);
            addDataTablesRow(options, data.Data, info.length, $table);
        }
    });
}
