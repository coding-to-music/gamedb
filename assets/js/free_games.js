if ($('#free-games-page').length > 0) {

    const table = $('table.table-datatable2');

    $('#types input:checkbox').change(function () {

        table.DataTable().draw();

    });

    table.DataTable($.extend(true, {}, dtDefaultOptions, {
        "ajax": function (data, callback, settings) {

            delete data.columns;
            delete data.length;
            delete data.search.regex;

            data.search.types = $('#types input:checkbox:checked').map(function () {
                return $(this).val();
            }).get();

            $.ajax({
                url: $(this).attr('data-path'),
                data: data,
                success: callback,
                dataType: 'json',
                cache: true
            });
        },
        "order": [[1, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-id', data[0]);
            $(row).attr('data-link', data[7]);
        },
        "columnDefs": [
            // Icon / Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<img src="' + row[2] + '" class="rounded square"><span>' + row[1] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img')
                }
            },
            // Score
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[3] + '%';
                }
            },
            // Type
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[4];
                },
                "orderable": false
            },
            // Platform
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[5];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('platforms platforms-align')
                },
                "orderable": false
            },
            // Install link
            {
                "targets": 4,
                "render": function (data, type, row) {
                    return '<a href="' + row[6] + '">Install</a>';
                },
                "orderable": false
            }
        ]
    }));
}
