if ($('#upcoming-page').length > 0) {

    $('table.table').gdbTable({
        searchFields: [
            $('#search'),
        ],
        tableOptions: {
            'order': [[1, 'desc']],
            'createdRow': function (row, data, dataIndex) {
                $(row).attr('data-app-id', data[0]);
                $(row).attr('data-link', data[3]);
            },
            'drawCallback': function (settings) {
                const api = this.api();
                if (api.order()[0] && api.order()[0][0] === 2) {
                    const rows = api.rows({page: 'current'}).nodes();

                    let last = null;
                    api.rows().every(function (rowIdx, tableLoop, rowLoop) {
                        let group = this.data()[6];
                        if (last !== group) {
                            $(rows).eq(rowIdx).before(
                                '<tr class="table-success"><td colspan="6">' + group + '</td></tr>',
                            );
                            last = group;
                        }
                    });
                }
            },
            'columnDefs': [
                // Icon / App Name
                {
                    'targets': 0,
                    'render': function (data, type, row) {
                        return '<a href="' + row[3] + '" class="icon-name"><div class="icon"><img data-lazy="' + row[2] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></a>';
                    },
                    'createdCell': function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    'orderable': false,
                },
                // Followers
                {
                    'targets': 1,
                    'render': function (data, type, row) {
                        return row[7];
                    },
                    'createdCell': function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    'orderSequence': ['desc'],
                },
                // Time
                {
                    'targets': 2,
                    'render': function (data, type, row) {
                        return '<span data-toggle="tooltip" data-placement="left" title="' + row[10] + '" data-livestamp="' + row[9] + '"></span>';
                    },
                    'createdCell': function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    'orderSequence': ['asc', 'desc'],
                },
                // Link
                {
                    'targets': 3,
                    'render': function (data, type, row) {
                        if (row[8]) {
                            return '<a href="' + row[8] + '" target="_blank" rel="noopener"><i class="fas fa-link"></i></a>';
                        }
                        return '';
                    },
                    'orderable': false,
                },
            ],
        },
    });
}
