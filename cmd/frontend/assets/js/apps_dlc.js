if ($('#apps-dlc-page').length > 0) {

    $('table.table').gdbTable({
        searchFields: [
            $('#search'),
        ],
        tableOptions: {
            'order': [[0, 'asc']],
            'createdRow': function (row, data, dataIndex) {
                $(row).attr('data-app-id', data[0]);
                $(row).attr('data-link', data[3]);
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
                    'orderSequence': ['asc', 'desc'],
                },
                // DLC
                {
                    'targets': 1,
                    'render': function (data, type, row) {
                        return row[4].toLocaleString();
                    },
                    'orderSequence': ['desc', 'asc'],
                },
                // Owned
                {
                    'targets': 2,
                    'render': function (data, type, row) {
                        return row[5].toLocaleString();
                    },
                    'createdCell': function (td, cellData, rowData, row, col) {
                        const p = rowData[5] / rowData[4] * 100;
                        $(td).css({
                            'background': 'linear-gradient(to right, rgba(0,0,0,.15) ' + p + '%, transparent ' + p + '%)',
                            'border-right': 'solid 1px rgba(0, 0, 0, 0.15)',
                        });
                    },
                    'orderable': false,
                },
                // Link
                {
                    'targets': 3,
                    'render': function (data, type, row) {
                        if (row[6]) {
                            return '<a href="' + row[6] + '" target="_blank" rel="noopener"><i class="fas fa-link"></i></a>';
                        }
                        return '';
                    },
                    'orderable': false,
                },
            ],
        },
    });
}
