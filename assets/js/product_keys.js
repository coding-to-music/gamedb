if ($('#product-keys-page').length > 0) {

    const $table = $('table.table-datatable2');

    // Setup datatable
    $table.DataTable($.extend(true, {}, dtDefaultOptions, {
        "ajax": function (data, callback, settings) {

            delete data.columns;
            delete data.length;
            delete data.search.regex;

            data.search.key = $('#key').val();
            data.search.value = $('#value').val();
            data.search.type = $('#type').val();

            $.ajax({
                url: $(this).attr('data-path'),
                data: data,
                success: callback,
                dataType: 'json',
                cache: true
            });
        },
        "order": [[0, 'asc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-id', data[0]);
            $(row).attr('data-link', data[3]);
        },
        "columnDefs": [
            // Icon / Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<img src="' + row[2] + '" class="rounded square" alt="' + row[1] + '"><span>' + row[1] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                }
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
    }));
}
