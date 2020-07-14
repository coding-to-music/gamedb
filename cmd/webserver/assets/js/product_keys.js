if ($('#product-keys-page').length > 0) {

    const searchOptions = {
        "order": [[0, 'asc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-app-id', data[0]);
            $(row).attr('data-link', data[3]);
        },
        "columnDefs": [
            // Icon / App Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<a href="' + row[3] + '" class="icon-name"><div class="icon"><img src="' + row[2] + '" alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></a>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                },
            },
            // Value
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[4];
                },
                "orderable": false
            },
        ]
    };

    const dt = $('#search-table').gdbTable({
        tableOptions: searchOptions,
        searchFields: [
            $('#key'),
            $('#value'),
            $('input[name=type]'),
        ],
    });

    $('#search button').on('click', function (e) {
        dt.draw();
    })

    const $tableSearch = $('#table-search');

    $('#apps-table').gdbTable({
        searchFields: [
            $tableSearch,
        ]
    });

    $('#packages-table').gdbTable({
        searchFields: [
            $tableSearch,
        ]
    });

    //
    $('#apps-table tbody tr').on('click', function (e) {

        $('#key').val($(this).attr('data-key'));
        $('input[name=type][value=apps]').prop("checked", true);
        $('a.nav-link[href="#search"]').tab('show');
    })

    $('#packages-table tbody tr').on('click', function (e) {

        $('#key').val($(this).attr('data-key'));
        $('input[name=type][value=packages]').prop("checked", true);
        $('a.nav-link[href="#search"]').tab('show');
    })
}
