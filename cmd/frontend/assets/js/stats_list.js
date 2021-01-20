if ($('#stats-list-page').length > 0) {

    (function ($, window) {
        'use strict';

        const options = {
            "order": [[1, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-link', data[0]);
            },
            "columnDefs": [
                // Icon / Stat Name
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        row[1] = row[1] ? row[1] : '-';
                        return '<i class="fas fa-star"></i> <span class="markable">' + row[1] + '</span>';
                    },
                    "orderSequence": ['asc', 'desc'],
                },
                // Apps
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return row[2].toLocaleString();
                    },
                    "orderSequence": ['desc', 'asc'],
                },
                // Apps Percent
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[6].toFixed(2).toLocaleString() + '%';
                    },
                    "orderSequence": ['desc', 'asc'],
                },
                // Price
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        if (row[2]) {
                            return row[3];
                        }
                        return '-';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderSequence": ['desc', 'asc'],
                },
                // Score
                {
                    "targets": 4,
                    "render": function (data, type, row) {
                        if (row[2]) {
                            return row[5];
                        }
                        return '-';
                    },
                    "orderSequence": ['desc', 'asc'],
                },
                // Players
                {
                    "targets": 5,
                    "render": function (data, type, row) {
                        if (row[2]) {
                            return row[4].toLocaleString();
                        }
                        return '-';
                    },
                    "orderSequence": ['desc', 'asc'],
                },
                // Max Discount
                {
                    "targets": 6,
                    "render": function (data, type, row) {
                        return row[7] + '%';
                    },
                    "orderSequence": ['desc', 'asc'],
                    "visible": window.location.pathname.includes("publishers") || window.location.pathname.includes("developers"),
                },
            ]
        };

        $('table.table').gdbTable({
            tableOptions: options,
            searchFields: [
                $('#search'),
                $('#type'),
            ]
        });

    })(jQuery, window);
}
