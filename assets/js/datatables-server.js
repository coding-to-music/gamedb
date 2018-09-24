$('table.table-datatable2')
    .on('page.dt', function () {
        $(this).fadeTo(100, 0.3);
    })
    .on('draw.dt', function () {
        $(this).fadeTo(100, 1);
    });

var defaultOptions = {
    "ajax": function (data, callback, settings) {

        delete data.columns;
        delete data.length;

        $.ajax({
            url: $(this).attr('data-path'),
            data: data,
            success: function (rdata) {
                callback(rdata);
            },
            dataType: 'json',
            cache: true
        });

    },
    "processing": false,
    "serverSide": true,
    "pageLength": 100,
    "fixedHeader": true,
    "paging": true,
    "ordering": true,
    "info": false,
    "searching": true,
    "autoWidth": false,
    "lengthChange": false,
    "stateSave": false,
    "orderMulti": false,
    "dom": 'r<"dt-pagination"p>t',
    "language": {
        "processing": '<i class="fas fa-spinner fa-spin fa-3x fa-fw"></i>'
    },
    "drawCallback": function (settings, json) {
        $(".paginate_button > a").on("focus", function () {
            $(this).blur(); // Fixes scrolling to pagination on every click
        });
    }
};

// Free Games Page
$('#free-games-page table.table-datatable2').DataTable($.extend(true, {}, defaultOptions, {
    "order": [[1, 'desc']],
    "createdRow": function (row, data, dataIndex) {
        $(row).attr('data-id', data[0]);
        $(row).attr('data-link', data[7]);
    },
    "columnDefs": [
        {
            "targets": 0,
            "render": function (data, type, row) {
                return '<img src="' + row[2] + '" class="rounded square"><span>' + row[1] + '</span>';
            },
            "createdCell": function (td, cellData, rowData, row, col) {
                $(td).addClass('img')
            }
        },
        {
            "targets": 1,
            "render": function (data, type, row) {
                return row[3] + '%';
            }
        },
        {
            "targets": 2,
            "render": function (data, type, row) {
                return row[4];
            }
        },
        {
            "targets": 3,
            "render": function (data, type, row) {
                return row[5];
            },
            "createdCell": function (td, cellData, rowData, row, col) {
                $(td).addClass('platforms platforms-align')
            },
            "orderable": false,
            "searchable": false
        },
        {
            "targets": 4,
            "render": function (data, type, row) {
                return '<a href="' + row[6] + '">Install</a>';
            },
            "orderable": false,
            "searchable": false
        }
    ]
}));

// Player Apps Tab
$('#player-page #games table.table-datatable2').DataTable($.extend(true, {}, defaultOptions, {
    "order": [[2, 'desc']],
    "createdRow": function (row, data, dataIndex) {
        $(row).attr('data-id', data[0]);
        $(row).attr('data-link', '/games/' + data[0]);
    },
    "columnDefs": [
        {
            "targets": 0,
            "render": function (data, type, row) {
                return '<img src="' + row[2] + '" class="rounded square"><span>' + row[1] + '</span>';
            },
            "createdCell": function (td, cellData, rowData, row, col) {
                $(td).addClass('img')
            }
        },
        {
            "targets": 1,
            "render": function (data, type, row) {
                return '$' + row[5];
            },
            "createdCell": function (td, cellData, rowData, row, col) {
                $(td).attr('data-sort', row[5])
            }
        },
        {
            "targets": 2,
            "render": function (data, type, row) {
                return row[4];
            },
            "createdCell": function (td, cellData, rowData, row, col) {
                $(td).attr('nowrap', 'nowrap').attr('data-sort', row[3])
            }
        },
        {
            "targets": 3,
            "render": function (data, type, row) {
                return '$' + row[6];
            }
        }
    ]
}));

// Players table
$('#player-page #games table.table-datatable2').DataTable($.extend(true, {}, defaultOptions, {
    "order": [[2, 'desc']],
    "createdRow": function (row, data, dataIndex) {
        $(row).attr('data-id', data[0]);
        $(row).attr('data-link', '/games/' + data[0]);
    },
    "columnDefs": [
        {
            "targets": 0,
            "render": function (data, type, row) {
                return row[4];
            },
            "createdCell": function (td, cellData, rowData, row, col) {
                $(td).addClass('font-weight-bold')
            }
        },
        {
            "targets": 1,
            "render": function (data, type, row) {
                return '<img src="' + row[2] + '" class="rounded square"><span>' + row[1] + '</span>';
            },
            "createdCell": function (td, cellData, rowData, row, col) {
                $(td).addClass('img')
            }
        },
        {
            "targets": 2,
            "render": function (data, type, row) {
                if (row[2]) {
                    return '<img data-toggle="tooltip" data-placement="left" title="' + row[1] + '" src="' + row[1] + '" class="rounded">';
                }
                return '';
            },
            "createdCell": function (td, cellData, rowData, row, col) {
                $(td).addClass('img')
            }
        },
        {
            "targets": 3,
            "render": function (data, type, row) {
                return '<img src="' + row[2] + '" class="rounded square"><span>' + row[1] + '</span>';
            },
            "createdCell": function (td, cellData, rowData, row, col) {
                $(td).addClass('img')
            }
        },
        {
            "targets": 4,
            "render": function (data, type, row) {
                return '$' + row[5];
            },
            "createdCell": function (td, cellData, rowData, row, col) {
                $(td).attr('data-sort', row[5])
            }
        },
        {
            "targets": x,
            "render": function (data, type, row) {
                return '$' + row[6];
            }
        },
        {
            "targets": x,
            "render": function (data, type, row) {
                return '$' + row[6];
            }
        },
        {
            "targets": x,
            "render": function (data, type, row) {
                return '$' + row[6];
            }
        },
        {
            "targets": x,
            "render": function (data, type, row) {
                return '$' + row[6];
            }
        },
        {
            "targets": x,
            "render": function (data, type, row) {
                return '$' + row[6];
            }
        }

    ]
}));
