if ($('#apps-page').length > 0) {

    var $chosens = $('select.form-control-chosen');

    $chosens.chosen({
        disable_search_threshold: 10,
        allow_single_deselect: true,
        rtl: false
    });

    $chosens.on('change', function (e) {

    });

    $chosens.trigger("chosen:updated");


    var $table = $('table.table-datatable2');

    $('#types input:checkbox').change(function () {

        $table.DataTable().draw();

    });

    $table.DataTable($.extend(true, {}, dtDefaultOptions, {
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
                }
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
