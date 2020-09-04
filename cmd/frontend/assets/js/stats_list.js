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
                // Price
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[3];
                    },
                    "orderSequence": ['desc', 'asc'],
                },
                // Score
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        return row[5];
                    },
                    "orderSequence": ['desc', 'asc'],
                },
                // Players
                {
                    "targets": 4,
                    "render": function (data, type, row) {
                        return row[4].toLocaleString();
                    },
                    "orderSequence": ['desc', 'asc'],
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
