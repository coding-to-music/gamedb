if ($('#apps-page').length > 0) {

    var $chosens = $('select.form-control-chosen');
    var $table = $('table.table-datatable2');

    // Setup drop downs
    $chosens.chosen({
        disable_search_threshold: 10,
        allow_single_deselect: true,
        rtl: false
    });

    $chosens.on('change', function (e) {
        $table.DataTable().draw();
    });

    // Setup datatable
    $table.DataTable($.extend(true, {}, dtDefaultOptions, {
        "ajax": function (data, callback, settings) {

            delete data.columns;
            delete data.length;
            delete data.search.regex;

            data.search.tags = $('#tags').val();
            data.search.genres = $('#genres').val();
            data.search.developers = $('#developers').val();
            data.search.publishers = $('#publishers').val();
            data.search.os = $('#os').val();
            data.search.types = $('#types').val();

            $.ajax({
                url: $(this).attr('data-path'),
                data: data,
                success: callback,
                dataType: 'json',
                cache: true
            });
        },
        "order": [[2, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-id', data[0]);
            $(row).attr('data-link', data[3]);
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
            // Type
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[4];
                },
                "orderable": false
            },
            // Score
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[5] + '%';
                }
            },
            // DLC Count
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[6];
                }
            }
        ]
    }));
}
