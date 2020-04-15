<template>
    <div class="layout">
        <Badge @click.native="publish(item)" v-for="item in Rooms" :key="item.StreamPath" :count="item.TsCount"
            class="room">
            <Icon type="ios-folder-open-outline" size="100" />
            <div class="size">{{unitFormat(item.TotalSize)}}</div>
            <div>{{item.StreamPath}}</div>
        </Badge>
        <div v-if="Rooms.length==0" class="empty">
            <Icon type="md-wine" size="50" />没有任何文件夹
        </div>
    </div>
</template>

<script>
export default {
    data() {
        return {
            Rooms: []
        };
    },
    methods: {
        fetchlist() {
            this.ajax.getJSON(this.apiHost + "/ts/list").then(x => {
                this.Rooms = x;
            });
        },
        publish(item) {
            this.$confirm("是否发布该目录", "提示").then(result => {
                if (result)
                    this.ajax
                        .getJSON(
                            this.apiHost +
                                "/ts/publish?streamPath=" +
                                item.StreamPath
                        )
                        .then(() => {
                            this.$toast.success("已发布" + item.StreamPath);
                        });
            });
        }
    },
    mounted() {
        this.fetchlist();
    }
};
</script>

<style>
.layout {
    padding-bottom: 30px;
    display: flex;
    flex-wrap: wrap;
    padding: 30px;
}
.room {
    text-align: center;
}
.room:hover {
    text-shadow: 0 0 12px #4199d0;
    border-radius: 10%;
    cursor: pointer;
}
.size {
    color: #ffc107;
    position: absolute;
    top: 40%;
    left: 0;
    right: 0;
}
</style>