if ($('#commits-page').length > 0) {

    const options = {
        "order": [[1, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-link', data[3]);
            $(row).attr('data-target', '_blank');
            if (data[4]) {
                $(row).addClass('table-success');
            }
        },
        "columnDefs": [
            // Message
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<div class="icon-name"><div class="name">' + row[0] + '</div></div>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('id', rowData[5]);
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false,
            },
            // Time
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return '<span data-toggle="tooltip" data-placement="left" title="' + row[6] + '" data-livestamp="' + row[1] + '"></span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false,
            },
            // Hash
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[5];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false,
            },
            // Deployed
            {
                "targets": 3,
                "render": function (data, type, row) {

                    if (page === null) {
                        page = $table.DataTable().page.info().page;
                    }

                    if (row[2] || page > 0) {
                        return '<i class="fas fa-check"></i>';
                    } else {
                        return '<i class="fas fa-times"></i>';
                    }
                },
                "orderable": false,
            }
        ]
    };

    let page = null;
    const $table = $('table.table');
    const dt = $table.gdbTable({tableOptions: options});

    dt.on('draw.dt', function (e, settings) {
        page = null;
    });
}
