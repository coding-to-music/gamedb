if ($('#product-keys-page').length > 0) {

    const $type = $('#type')
    const $key = $('#key');
    const $comparator = $('#comparator');
    const $value = $('#value');

    // Setup drop downs
    $type.chosen({
        allow_single_deselect: false,
        disable_search_threshold: 10,
    });

    $key.chosen({
        allow_single_deselect: false,
    });

    $comparator.chosen({
        allow_single_deselect: true,
        disable_search_threshold: 10,
    });

    $comparator.on('chosen:updated change', function (e) {
        if ($comparator.val()) {
            $('#value-wrapper').removeClass('d-none');
        } else {
            $('#value-wrapper').addClass('d-none');
        }
    })

    // Search results
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
        searchFields: [$type, $key, $comparator, $value],
    });

    $('#search button').on('click', function (e) {
        dt.draw();
    });

    $('#apps-table').gdbTable({
        searchFields: [
            $('#app-search'),
        ]
    });

    $('#packages-table').gdbTable({
        searchFields: [
            $('#package-search'),
        ]
    });

    //
    $('#apps-table tbody tr').on('click', function (e) {

        $key.val($(this).attr('data-key')).trigger("chosen:updated");
        $type.val('apps').trigger("chosen:updated");

        $('a.nav-link[href="#search"]').tab('show');
        dt.draw();
    });

    $('#packages-table tbody tr').on('click', function (e) {

        $key.val($(this).attr('data-key')).trigger("chosen:updated");
        $type.val('packages').trigger("chosen:updated");

        $('a.nav-link[href="#search"]').tab('show');
        dt.draw();
    });
}
