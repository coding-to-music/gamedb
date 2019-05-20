if ($('#groups-page').length > 0) {

    $('form').on('submit', function (e) {

        $table.DataTable().draw();
        return false;
    });

    $('table.table-datatable2').DataTable($.extend(true, {}, dtDefaultOptions, {
        "ajax": function (data, callback, settings) {

            delete data.columns;
            delete data.length;
            delete data.search;

            data.search = {};
            data.search.search = $('#search').val();

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
            $(row).attr('data-link', data[2]);
        },
        "columnDefs": [
            // Icon / Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<img data-src="/assets/img/no-app-image-square.jpg" src="' + row[3] + '" class="rounded square" alt="' + row[1] + '"><span>' + row[1] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false,
            },
            // Headline
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[4]
                },
                "orderable": false,
            },
            // Members
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[5].toLocaleString();
                },
                "orderable": false,
            },
            // Link
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return '<i class="fas fa-link" data-target="_blank" data-link="https://steamcommunity.com/groups/' + row[6] + '"></i>';
                },
                "orderable": false,
            },
        ]
    }));
}
