$('table.table-datatable2')
    .on('page.dt', function () {
        $(this).fadeTo(100, 0.3);
        $('html, body').animate({scrollTop: 0}, 500);
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
    "dom": 't<"dt-pagination"p>',
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
            }
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
            "orderable": false,
            "searchable": false
        },
        // Install link
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
        // Price
        {
            "targets": 1,
            "render": function (data, type, row) {
                return '$' + row[5];
            }
        },
        // Time
        {
            "targets": 2,
            "render": function (data, type, row) {
                return row[4];
            },
            "createdCell": function (td, cellData, rowData, row, col) {
                $(td).attr('nowrap', 'nowrap');
            }
        },
        // Price/Time
        {
            "targets": 3,
            "render": function (data, type, row) {
                return '$' + row[6];
            }
        }
    ]
}));

// Players table
$('#ranks-page table.table-datatable2').DataTable($.extend(true, {}, defaultOptions, {
    "order": [[3, 'asc']],
    "createdRow": function (row, data, dataIndex) {
        $(row).attr('data-id', data[1]);
        $(row).attr('data-link', '/players/' + data[1]);
    },
    "columnDefs": [
        // Rank
        {
            "targets": 0,
            "render": function (data, type, row) {
                return row[0];
            },
            "createdCell": function (td, cellData, rowData, row, col) {
                $(td).addClass('font-weight-bold')
            },
            "orderable": false
        },
        // Player
        {
            "targets": 1,
            "render": function (data, type, row) {
                return '<img src="' + row[3] + '" class="rounded square"><span>' + row[2] + '</span>';
            },
            "createdCell": function (td, cellData, rowData, row, col) {
                $(td).addClass('img')
            },
            "orderable": false
        },
        // Flag
        {
            "targets": 2,
            "render": function (data, type, row) {
                if (row[11]) {
                    return '<img data-toggle="tooltip" data-placement="top" title="' + row[12] + '" src="' + row[11] + '" class="rounded">';
                }
                return '';
            },
            "createdCell": function (td, cellData, rowData, row, col) {
                $(td).addClass('img');
            },
            "orderable": false
        },
        // Avatar 2 / Level
        {
            "targets": 3,
            "render": function (data, type, row) {
                return '<div class="' + row[4] + ' square"></div><span>' + row[5].toLocaleString() + '</span>';
            },
            "createdCell": function (td, cellData, rowData, row, col) {
                $(td).addClass('img')
            }
        },
        // Games
        {
            "targets": 4,
            "render": function (data, type, row) {
                return row[6].toLocaleString();
            }
        },
        // Badges
        {
            "targets": 5,
            "render": function (data, type, row) {
                return row[7].toLocaleString();
            }
        },
        // Time
        {
            "targets": 6,
            "render": function (data, type, row) {
                return row[8];
            },
            "createdCell": function (td, cellData, rowData, row, col) {
                $(td).attr('nowrap', 'nowrap').attr('data-toggle', 'tooltip').attr('data-placement', 'top').attr('title', rowData[9]);
            }
        },
        // Friends
        {
            "targets": 7,
            "render": function (data, type, row) {
                return row[10].toLocaleString();
            }
        }
    ]
}));
